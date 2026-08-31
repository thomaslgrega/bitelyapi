# Recipe Images upload direct to R2; the API signs and records but never touches bytes

A Shared Recipe carries one image. It is uploaded when the Recipe is shared, stored in a Cloudflare R2 bucket, and served from that bucket's public URL. The API mints a presigned upload URL and records the resulting key. It never sees the bytes in either direction.

**One image, uploaded at share time.** A Private Recipe's image stays on the device, the same way the Recipe does — the server storing it would make Bitely a sync backend for data the corpus is defined not to hold. One image serves both the card and the hero. Two derivatives would double the upload, the failure modes and the orphan surface to solve a problem the corpus is too small to have; when it does have it, Cloudflare Images transforms serve `?width=` off the same stored object, which is why the column holds a key rather than a URL.

**The API never touches bytes.** Uploads go straight from the client to R2 through a presigned PUT; reads come from the bucket's public URL. R2 charges no egress, so keeping bytes off the API is most of the reason to pick R2 at all — proxying them through a Render container would spend the container's bandwidth and hold a request goroutine open for the length of a phone upload on a cell connection. The read path is public because `GET /recipes/{id}` and the Feed already are: guarding an image URL would guard nothing the Recipe row does not hand out first.

**Staged, then promoted.** A presigned URL is issued against `incoming/<uuid>`, and sharing promotes the object to `recipes/<id>/image.jpg` with a server-side `CopyObject`. The promoted key is derived by the server from the Recipe id and is never supplied by the client, so a client can name nothing outside `incoming/`, and a key is validated for shape before it is used. A prefix-scoped lifecycle rule on `incoming/` deletes whatever is abandoned or rejected, which is why there is no reaper: the alternative — signing straight into the final key — needs a job that diffs the bucket against the table forever.

The order is copy, then row, then a best-effort delete of the staged object. A failure anywhere leaves an object the lifecycle rule eventually eats, rather than a Recipe row pointing at nothing.

**Enforcement is after the upload, not at signing.** R2 documents that a signed `Content-Type` is enforced: a mismatched upload fails with `403 SignatureDoesNotMatch`. It offers nothing equivalent for size. S3's mechanism is POST Policy with `content-length-range`, and R2 does not implement `POST` uploads at all; signing `Content-Length` pins an exact byte count rather than a ceiling, and Cloudflare does not document whether R2 validates it. So the size and type limits — 5MB, `image/jpeg` and `image/png` — are checked with a `HEAD` on the staged object before it is promoted. Nothing upstream of that can check them.

The cost is that an abusive upload is stored and billed until the lifecycle rule catches it, whose floor is one day and whose deletion lags up to twenty-four hours. Closing that would mean streaming uploads through a Worker, which gives up the direct-to-R2 property this decision exists for.

**Rows hold keys; responses hold URLs.** `recipes.image_key` names the object; the API composes `R2_PUBLIC_BASE_URL + key` when a row becomes a model. Storing the URL would write the hostname into every row and make a domain change or a move to Cloudflare Images a table rewrite. Composition happens at the repository's scan sites because that is the one boundary every row crosses on its way to a handler.

That hostname is R2's Public Development URL — a generated `pub-<hash>.r2.dev` — rather than a custom domain, because Bitely is a personal project and a custom domain means owning one and moving it onto Cloudflare. Cloudflare rate limits `r2.dev` and does not intend it for production traffic, so this is the decision to revisit first if the app ships. It costs one environment variable, which is the point of holding a key.

No image is the empty string, not `NULL`, so the column scans into a plain `string` like the columns beside it. The dev seed's picsum URLs are dropped rather than passed through the composer: a special case for a fixture would live in the read path permanently.

## Consequences

The request and the response name the image differently. A client sends `image_key` — a claim ticket for a staged upload — and receives `image_url`, a fetchable address. `CreateRecipeInput` and `Recipe` therefore stop mirroring each other field-for-field, and `Recipe` carries both fields because `PUT /recipes/{id}` decodes into the struct `GET /recipes/{id}` encodes. Resolving clears the key so a response never names the bucket layout.

Presigned URLs work on neither the Public Development URL nor a custom domain, so uploads address `<account-id>.r2.cloudflarestorage.com` directly. The read hostname and the write hostname are different by construction, and `R2_PUBLIC_BASE_URL` names only the former.

Images are public and effectively permanent. There is no moderation layer and the API cannot enforce anything about what an image depicts, only its size and declared type. Deleting or replacing a Recipe deletes its object best-effort; an R2 failure does not fail the request, because the Recipe genuinely did change. A rare orphan survives that, and is accepted rather than engineered against.

Two facts Cloudflare's documentation does not settle were tested against the bucket, and both came out in this decision's favour.

**R2 enforces the signed `Content-Length`.** A presigned URL minted for 1024 bytes answers `403` to a 4096-byte upload and stores nothing, the same way it refuses a mismatched content type. The size limit therefore holds at two points, not one: the signature caps what a given URL can write at all, and the `HEAD` before promotion is what the limit is actually specified against. The `HEAD` stays load-bearing regardless — an upload can still lie about its type by declaring `image/jpeg` and sending anything, which only reading the stored object catches.

That an oversized upload is refused outright is stronger than this decision assumed. It does not close the billing window above, because a signed URL can be replayed for its whole life, but it does mean a client cannot turn one signature into an object of arbitrary size.

**The direct `HEAD`/`CopyObject` path does not hit the checksum `501`.** A `HEAD` reports the stored size and type, the server-side copy to the derived key succeeds, and the copy carries its content type across. `RequestChecksumCalculation` and `ResponseChecksumValidation` are set to `WhenRequired` anyway: nothing here needs a checksum, and the setting costs nothing.

Both were established with a build-tagged tier in `internal/r2` that talks to a real bucket (`go test -tags manual ./internal/r2/`). It is tagged out of `go test ./...`, which runs with neither database nor network.

Presigning itself needs no checksum configuration, and must not set `ChecksumAlgorithm` on the input: doing so signs `x-amz-checksum-crc32` for an empty body and every real upload through the URL fails.
