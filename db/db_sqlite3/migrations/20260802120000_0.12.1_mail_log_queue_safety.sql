-- +goose Up
DELETE FROM "mail_logs"
WHERE EXISTS (
    SELECT 1 FROM "results"
    WHERE "results"."campaign_id" = "mail_logs"."campaign_id"
      AND "results"."r_id" = "mail_logs"."r_id"
      AND "results"."status" = 'Error'
);

DELETE FROM "mail_logs"
WHERE "id" NOT IN (
    SELECT MIN("id") FROM "mail_logs" GROUP BY "campaign_id", "r_id"
);

CREATE UNIQUE INDEX "ux_mail_logs_campaign_rid" ON "mail_logs" ("campaign_id", "r_id");
CREATE INDEX "ix_mail_logs_due" ON "mail_logs" ("processing", "send_date");

-- +goose Down
DROP INDEX "ix_mail_logs_due";
DROP INDEX "ux_mail_logs_campaign_rid";
