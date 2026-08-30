# Recipe Images upload direct to R2; the API signs and records but never touches bytes

A Shared Recipe carries one image. It is uploaded when the Recipe is shared, stored in a Cloudflare R2 bucket, and served from a public Cloudflare custom domain. The API mints a presigned upload URL and records the resulting key. It never sees the bytes in either direction.

**One image, uploaded at share time.** A Private Recipe's image stays on the device, the same way the Recipe does — the server storing it would make Bitely a sync backend for data the corpus is defined not to hold. One image serves both the card and the hero. Two derivatives would double the upload, the failure modes and the orphan surface to solve a problem the corpus is too small to have; when it does have it, Cloudflare Images transforms serve `?width=` off the same stored object, which is why the column holds a key rather than a URL.

**The API never touches bytes.** Uploads go straight from the client to R2 through a presigned PUT; reads come from the public domain. R2 charges no egress, so keeping bytes off the API is most of the reason to pick R2 at all — proxying them through a Render container would spend the container's bandwidth and hold a request goroutine open for the length of a phone upload on a cell connection. The read path is public because `GET /recipes/{id}` and the Feed already are: guarding an image URL would guard nothing the Recipe row does not hand out first.

**Staged, then promoted.** A presigned URL is issued against `incoming/<uuid>`, and sharing promotes the object to `recipes/<id>/image.jpg` with a server-side `CopyObject`. The promoted key is derived by the server from the Recipe id and is never supplied by the client, so a client can name nothing outside `incoming/`, and a key is validated for shape before it is used. A prefix-scoped lifecycle rule on `incoming/` deletes whatever is abandoned or rejected, which is why there is no reaper: the alternative — signing straight into the final key — needs a job that diffs the bucket against the table forever.

The order is copy, then row, then a best-effort delete of the staged object. A failure anywhere leaves an object the lifecycle rule eventually eats, rather than a Recipe row pointing at nothing.

**Enforcement is after the upload, not at signing.** R2 documents that a signed `Content-Type` is enforced: a mismatched upload fails with `403 SignatureDoesNotMatch`. It offers nothing equivalent for size. S3's mechanism is POST Policy with `content-length-range`, and R2 does not implement `POST` uploads at all; signing `Content-Length` pins an exact byte count rather than a ceiling, and Cloudflare does not document whether R2 validates it. So the size and type limits — 5MB, `image/jpeg` and `image/png` — are checked with a `HEAD` on the staged object before it is promoted. Nothing upstream of that can check them.

The cost is that an abusive upload is stored and billed until the lifecycle rule catches it, whose floor is one day and whose deletion lags up to twenty-four hours. Closing that would mean streaming uploads through a Worker, which gives up the direct-to-R2 property this decision exists for.

**Rows hold keys; responses hold URLs.** `recipes.image_key` names the object; the API composes `R2_PUBLIC_BASE_URL + key` when a row becomes a model. Storing the URL would write the CDN hostname into every row and make a domain change or a move to Cloudflare Images a table rewrite. Composition happens at the repository's scan sites because that is the one boundary every row crosses on its way to a handler.

No image is the empty string, not `NULL`, so the column scans into a plain `string` like the columns beside it. The dev seed's picsum URLs are dropped rather than passed through the composer: a special case for a fixture would live in the read path permanently.

## Consequences

The request and the response name the image differently. A client sends `image_key` — a claim ticket for a staged upload — and receives `image_url`, a fetchable address. `CreateRecipeInput` and `Recipe` therefore stop mirroring each other field-for-field, and `Recipe` carries both fields because `PUT /recipes/{id}` decodes into the struct `GET /recipes/{id}` encodes. Resolving clears the key so a response never names the bucket layout.

Images are public and effectively permanent. There is no moderation layer and the API cannot enforce anything about what an image depicts, only its size and declared type. Deleting or replacing a Recipe deletes its object best-effort; an R2 failure does not fail the request, because the Recipe genuinely did change. A rare orphan survives that, and is accepted rather than engineered against.

Two facts could not be settled from Cloudflare's documentation and are recorded here once tested against a real bucket:

- Whether R2 rejects an upload whose size differs from a signed `Content-Length`. Cloudflare documents the content-type case and only that case.
- Whether the direct `HEAD`/`CopyObject` path hits the checksum `501` that affected non-presigned `PutObject` from `aws-sdk-go-v2` `service/s3` v1.73.0. It appears fixed server-side; the opt-out is `RequestChecksumCalculation` and `ResponseChecksumValidation` set to `WhenRequired`.

Presigning itself needs no checksum configuration, and must not set `ChecksumAlgorithm` on the input: doing so signs `x-amz-checksum-crc32` for an empty body and every real upload through the URL fails.
