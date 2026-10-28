-- The payments removed by the up migration can't be restored.
ALTER TABLE payments DROP COLUMN IF EXISTS shop_id;
