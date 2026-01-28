9.  **Soft Deletes:** A `deleted_at` column support so data isn't permanently lost immediately.
10. **Webhooks / Events:** A system to notify external services when data changes (e.g., "Order Created" -> Trigger Email).
7.  **Batch Operations:** Bulk `Create`, `Update`, or `Delete` APIs. Creating 100 products one by one is slow.
8.  **File Uploads / Media Handling:** No visible support for `multipart/form-data` to upload images/files, which is mandatory for CMS/E-com.
6.  **Aggregation:** Endpoints for `count`, `sum`, `avg`, `min`, `max`. (Critical for dashboards/analytics).
5.  **Relations / Population:** Fetching related data (e.g., "Get Product **and** its Author"). Currently requires multiple API calls.