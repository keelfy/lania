INSERT INTO name_prefixes (id, name, metadata) VALUES 
  ('a2fcbcae-d317-4bb5-91da-5f4605e5ca1f', 'Meowdy', '{"prefix": "", "image": "https://czx1jtlf2o.ufs.sh/f/0UHiIrRo6i8cpsO2dlw401exSHy3L6fZpuANMjKYwXV9dqvh"}'),
  ('429917ba-6ed8-4638-8803-e07150586717', 'Cat', '{"prefix": "", "image": "https://czx1jtlf2o.ufs.sh/f/0UHiIrRo6i8clv8rVKF6tIyhSxkCEQcwFgboHBdU7L4p8afD"}'),
  ('b213bfe8-dfcd-452f-af60-1890455a2c6b', 'HelloKittySkull', '{"prefix": "", "image": "https://czx1jtlf2o.ufs.sh/f/0UHiIrRo6i8cFzZKxHIjYlnvfiUrumoS0gcDQqX98pE2BbZI"}'),
  ('ed65c015-786f-4ce7-99f5-b30b5bf5e3f7', 'CatSus', '{"prefix": "", "image": "https://czx1jtlf2o.ufs.sh/f/0UHiIrRo6i8cbxmYR39jT7i9h3X0yLRptdlZCvF5VwOSKfU8"}'),
  ('3f869c5a-4bef-4511-8a49-9ddb590ccbc9', 'Cute', '{"prefix": "", "image": "https://czx1jtlf2o.ufs.sh/f/0UHiIrRo6i8c4PdMHtSh857rSIHFZd4Cc9QnJTBfvtqGKhRw", "noSpace": true}');

INSERT INTO products (id, category, price_name, metadata) VALUES 
  ('d1dea961-a68e-4c8b-af35-f6306eead8d5', 'name-prefix', 'name_prefix', '{"prefix": "https://czx1jtlf2o.ufs.sh/f/0UHiIrRo6i8cpsO2dlw401exSHy3L6fZpuANMjKYwXV9dqvh", "namePrefixId": "a2fcbcae-d317-4bb5-91da-5f4605e5ca1f"}'),
  ('d775ec0a-0efe-4eb2-b5c2-55904b839d12', 'name-prefix', 'name_prefix', '{"prefix": "https://czx1jtlf2o.ufs.sh/f/0UHiIrRo6i8clv8rVKF6tIyhSxkCEQcwFgboHBdU7L4p8afD", "namePrefixId": "429917ba-6ed8-4638-8803-e07150586717"}'),
  ('deb2f617-7258-4f3f-8111-76be6e2809e9', 'name-prefix', 'name_prefix', '{"prefix": "https://czx1jtlf2o.ufs.sh/f/0UHiIrRo6i8cFzZKxHIjYlnvfiUrumoS0gcDQqX98pE2BbZI", "namePrefixId": "b213bfe8-dfcd-452f-af60-1890455a2c6b"}'),
  ('86bd3132-edc0-453b-8fb8-a7c66dc20cff', 'name-prefix', 'name_prefix', '{"prefix": "https://czx1jtlf2o.ufs.sh/f/0UHiIrRo6i8cbxmYR39jT7i9h3X0yLRptdlZCvF5VwOSKfU8", "namePrefixId": "ed65c015-786f-4ce7-99f5-b30b5bf5e3f7"}'),
  ('6609f7e4-26f3-4074-a6c3-ecce6cdae65a', 'name-prefix', 'name_prefix', '{"prefix": "https://czx1jtlf2o.ufs.sh/f/0UHiIrRo6i8c4PdMHtSh857rSIHFZd4Cc9QnJTBfvtqGKhRw", "namePrefixId": "3f869c5a-4bef-4511-8a49-9ddb590ccbc9"}');

INSERT INTO product_localizations (product_id, locale, name, description) VALUES 
  ('d1dea961-a68e-4c8b-af35-f6306eead8d5', 'en', 'Meowdy', 'Icon that appears before your name'),
  ('d775ec0a-0efe-4eb2-b5c2-55904b839d12', 'en', 'Cat', 'Icon that appears before your name'),
  ('deb2f617-7258-4f3f-8111-76be6e2809e9', 'en', 'HelloKittySkull', 'Icon that appears before your name'),
  ('86bd3132-edc0-453b-8fb8-a7c66dc20cff', 'en', 'CatSus', 'Icon that appears before your name'),
  ('6609f7e4-26f3-4074-a6c3-ecce6cdae65a', 'en', 'Cute', 'Icon that appears before your name');

INSERT INTO product_localizations (product_id, locale, name, description) VALUES 
  ('d1dea961-a68e-4c8b-af35-f6306eead8d5', 'ru', 'Meowdy', 'Иконка, которая будет отображаться перед вашим именем'),
  ('d775ec0a-0efe-4eb2-b5c2-55904b839d12', 'ru', 'Cat', 'Иконка, которая будет отображаться перед вашим именем'),
  ('deb2f617-7258-4f3f-8111-76be6e2809e9', 'ru', 'HelloKittySkull', 'Иконка, которая будет отображаться перед вашим именем'),
  ('86bd3132-edc0-453b-8fb8-a7c66dc20cff', 'ru', 'CatSus', 'Иконка, которая будет отображаться перед вашим именем'),
  ('6609f7e4-26f3-4074-a6c3-ecce6cdae65a', 'ru', 'Cute', 'Иконка, которая будет отображаться перед вашим именем');

INSERT INTO ed_products (product_id, ed_product_id) VALUES 
  ('d1dea961-a68e-4c8b-af35-f6306eead8d5', 1022187),
  ('d775ec0a-0efe-4eb2-b5c2-55904b839d12', 1022188),
  ('deb2f617-7258-4f3f-8111-76be6e2809e9', 1022186),
  ('86bd3132-edc0-453b-8fb8-a7c66dc20cff', 1022189),
  ('6609f7e4-26f3-4074-a6c3-ecce6cdae65a', 1022190);

INSERT INTO name_prefixes (id, name, metadata) VALUES 
  ('f158fbb3-4492-4118-a063-0a317d465aa9', 'Yobshestvo', '{"prefix": "<hover:show_text:''Участник Ёбщества''><shadow:yellow><reset>", "image": "https://czx1jtlf2o.ufs.sh/f/0UHiIrRo6i8c9J5C8OLu6FuCQkw4zHpxVNvnJIi1g8LSyWZj"}');