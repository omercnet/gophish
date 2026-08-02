-- +goose Up
DELETE ml FROM `mail_logs` ml
JOIN `results` r
  ON r.`campaign_id` = ml.`campaign_id` AND r.`r_id` = ml.`r_id`
WHERE r.`status` = 'Error';

DELETE duplicate FROM `mail_logs` duplicate
JOIN `mail_logs` keeper
  ON keeper.`campaign_id` = duplicate.`campaign_id`
 AND LEFT(keeper.`r_id`, 190) = LEFT(duplicate.`r_id`, 190)
 AND keeper.`id` < duplicate.`id`;

ALTER TABLE `mail_logs`
  ADD UNIQUE INDEX `ux_mail_logs_campaign_rid` (`campaign_id`, `r_id`(190)),
  ADD INDEX `ix_mail_logs_due` (`processing`, `send_date`);

-- +goose Down
ALTER TABLE `mail_logs`
  DROP INDEX `ix_mail_logs_due`,
  DROP INDEX `ux_mail_logs_campaign_rid`;
