# Recipe Images upload direct to R2; the API signs and records but never touches bytes

A Shared Recipe carries one image. It is uploaded when the Recipe is shared, stored in a Cloudflare R2 bucket, and served from that bucket's public URL. The API mints a presigned upload URL and records the resulting key. It never sees the bytes in either direction.

**One image, uploaded at share time.** A Private Recipe's image stays on the device, the same way the Recipe does — the server storing it would make Bitely a sync backend for data the corpus is defined not to hold. One image serves both the card and the hero. Two derivatives would double the upload, the failure modes and the orphan surface to solve a problem the corpus is too small to have; when it does have it, Cloudflare Images transforms serve `?width=` off the same stored object, which is why the column holds a key rather than a URL.

**The API never touches bytes.** Uploads go straight from the client to R2 through a presigned PUT; reads come from the bucket's public URL. R2 charges no egress, so keeping bytes off the API is most of the reason to pick R2 at all — proxying them through a Render container would spend the container's bandwidth and hold a request goroutine open for the length of a phone upload on a cell connection. The read path is public because `GET /recipes/{id}` and the Feed already are: guarding an image URL would guard nothing the Recipe row does not hand out first.

**Staged, then promoted.** A presigned URL is issued against `incoming/<uuid>`, and sharing promotes the object under `recipes/<id>/` with a server-side `CopyObject`. The promoted key is derived by the server from the Recipe id and is never supplied by the client, so a client can name nothing outside `incoming/`, and a key is validated for shape before it is used. A prefix-scoped lifecycle rule on `incoming/` deletes whatever is abandoned or rejected, which is why there is no reaper: the alternative — signing straight into the final key — needs a job that diffs the bucket against the table forever.

The promoted key is unique per upload — `recipes/<id>/<uuid>.jpg` rather than a fixed filename — so a replacement lands beside the object the Recipe currently serves instead of on top of it. A Recipe Image therefore changes when its row changes and at no other moment: a copy that succeeds before a row that fails leaves an orphan, not a Recipe serving an image its Author did not publish. The row is then the only record of which object is live, so the mutation that changes it reports the key it replaced, read under its own lock: two writes racing to replace the same image each discard what they actually superseded, rather than both discarding what a read before them saw.

**The signature bounds one upload; the `HEAD` is what the limits are checked against.** R2 enforces both signed headers: a mismatched `Content-Type` and a mismatched `Content-Length` each fail with `403 SignatureDoesNotMatch`, which the Consequences below record as tested. A given URL therefore cannot be turned into an object of some other size, but that is a property of each signature rather than a limit the system enforces — signing pins one exact byte count, never a ceiling, and S3's mechanism for a range, POST Policy with `content-length-range`, does not exist on R2, which implements no `POST` uploads at all.

So the limits themselves — 5MB, `image/jpeg` and `image/png` — are checked with a `HEAD` on the staged object before it is promoted. That check is not redundant: the signature constrains what the bytes are declared to be, and only reading the stored object says what they are. An upload can still declare `image/jpeg` and send anything at all.

The cost is that a rejected upload is stored and billed until the lifecycle rule catches it, whose floor is one day and whose deletion lags up to twenty-four hours. Closing that would mean streaming uploads through a Worker, which gives up the direct-to-R2 property this decision exists for.

**Rows hold keys; responses hold URLs.** `recipes.image_key` names the object; the API composes `R2_PUBLIC_BASE_URL + key` when a row becomes a model. Storing the URL would write the hostname into every row and make a domain change or a move to Cloudflare Images a table rewrite. Composition happens at the repository's scan sites because that is the one boundary every row crosses on its way to a handler.

That hostname is R2's Public Development URL — a generated `pub-<hash>.r2.dev` — rather than a custom domain, because Bitely is a personal project and a custom domain means owning one and moving it onto Cloudflare. Cloudflare rate limits `r2.dev` and does not intend it for production traffic, so this is the decision to revisit first if the app ships. It costs one environment variable, which is the point of holding a key.

No image is the empty string, not `NULL`, so the column scans into a plain `string` like the columns beside it. The dev seed's picsum URLs are dropped rather than passed through the composer: a special case for a fixture would live in the read path permanently.

## Consequences

The request and the response name the image differently. A client sends `image_key` — a claim ticket for a staged upload — and receives `image_url`, a fetchable address. `CreateRecipeInput` and `Recipe` therefore stop mirroring each other field-for-field.

**A Recipe Image is written through its own sub-resource.** `PUT /recipes/{id}/image` takes a staged key and answers `200` with the `image_url` the promotion produced, because the promoted key is minted by the server and a client that got `204` could only learn where its photo landed by fetching the Recipe again. `DELETE /recipes/{id}/image` answers `204` whether or not there was an image to remove, so a retried save cannot fail on its second attempt. `PUT /recipes/{id}` carries no `image_key` and refuses one with a `400` naming the endpoint that does.

A recipe write replaces every field it carries, so a field a client omits is a field it deletes. That rule is safe for text, which the client holds and can send again, and unsafe for an image, whose bytes this decision deletes on replacement and whose key a client is never handed. Taking the image out of the recipe write is what keeps the rule literally true for every field that write still carries: an unchanged image is expressed by not calling the image endpoint, the one representation an omission cannot corrupt.

`POST /recipes` keeps `image_key`. The promoted key derives from the Recipe id, which does not exist until create runs, so there is nothing for a separate image write to address yet; splitting create in two would open a window where a Shared Recipe has no image and add a failure mode to the path every share takes.

Neither image endpoint promotes before `GetRecipeAuthor` has answered, for the reason the recipe write already establishes ownership first: promotion writes to a key derived from the path id, so leaving ownership to the row update would let a stranger's staged upload reach an Author's Recipe. A `404` on these routes means the Recipe is missing or belongs to someone else and nothing besides, which is why an imageless `DELETE` answers `204` instead.

A client changing both text and image issues two writes, and they do not commit together — R2 and Postgres cannot, which is why a row update that fails after a promotion already leaves an object behind. The image write goes first and the client abandons the save if it fails: the irreversible half then either happened or did not, and a retry costs nothing.

**Responses hold no key, but they conceal none either.** `image_url` is `R2_PUBLIC_BASE_URL + "/" + key` and that base is one constant every client holds, so any URL yields its key by removing a prefix. An earlier version of this decision claimed resolving kept a response from naming the bucket layout; it never did. What keeps `image_key` off responses is that the read-side and write-side values are different kinds of thing — a promoted `recipes/<id>/<uuid>.jpg` against a staged `incoming/<uuid>` — and the write path refuses the former deliberately. `Recipe` is decoded by nothing now that `PUT` has its own input type, so the key it scans into is tagged out of JSON rather than blanked on the way past.

Presigned URLs work on neither the Public Development URL nor a custom domain, so uploads address `<account-id>.r2.cloudflarestorage.com` directly. The read hostname and the write hostname are different by construction, and `R2_PUBLIC_BASE_URL` names only the former.

Images are public and effectively permanent. There is no moderation layer and the API cannot enforce anything about what an image depicts, only its size and declared type. Deleting or replacing a Recipe deletes its object best-effort; an R2 failure does not fail the request, because the Recipe genuinely did change. A rare orphan survives that, and is accepted rather than engineered against.

Two facts Cloudflare's documentation does not settle were tested against the bucket, and both came out in this decision's favour.

**R2 enforces the signed `Content-Length`.** A presigned URL minted for 1024 bytes answers `403` to a 4096-byte upload and stores nothing, the same way it refuses a mismatched content type. The size limit therefore holds at two points, not one: the signature caps what a given URL can write at all, and the `HEAD` before promotion is what the limit is actually specified against. The `HEAD` stays load-bearing regardless — an upload can still lie about its type by declaring `image/jpeg` and sending anything, which only reading the stored object catches.

That an oversized upload is refused outright is stronger than this decision assumed. It does not close the billing window above, because a signed URL can be replayed for its whole life, but it does mean a client cannot turn one signature into an object of arbitrary size.

**The direct `HEAD`/`CopyObject` path does not hit the checksum `501`.** A `HEAD` reports the stored size and type, the server-side copy to the derived key succeeds, and the copy carries its content type across. `RequestChecksumCalculation` and `ResponseChecksumValidation` are set to `WhenRequired` anyway: nothing here needs a checksum, and the setting costs nothing.

Both were established by hand against the bucket. Nothing in the suite talks to R2: `go test ./...` runs with neither database nor network.

Presigning itself needs no checksum configuration, and must not set `ChecksumAlgorithm` on the input: doing so signs `x-amz-checksum-crc32` for an empty body and every real upload through the URL fails.
