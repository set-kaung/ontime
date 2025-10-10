-- name: InsertReport :exec
INSERT INTO report (listing_id,reporter_id,datetime,report_reason,status,additional_detail)
VALUES ($1,$2,NOW(),$3,'pending',$4);
