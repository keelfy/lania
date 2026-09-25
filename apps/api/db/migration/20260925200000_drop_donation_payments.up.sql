-- DonationAlerts and FreeKassa are gone; EasyDonate is the only payment method.
-- oauth2_integrations held only the DonationAlerts token.
DROP TABLE IF EXISTS oauth2_integrations;

-- Every paid order granted access with the "freekassa" source, whatever the payment method was.
UPDATE profile_accesses SET source = 'order' WHERE source = 'freekassa';
