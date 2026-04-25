INSERT INTO products (id, category, price, metadata) VALUES
    ('63406c8d-67c2-43c1-a8a7-e7e79b7ac2c3', 'name-color', 199, '{"colors": ["#403B4A", "#E7E9BB"]}'),
    ('d9607e73-ad81-46e6-b0af-ee33b5987e16', 'name-color', 199, '{"colors": ["#000000", "#434343"]}'),
    ('2a65d174-8896-411a-a6dd-8341b3620171', 'name-color', 199, '{"colors": ["#ffe259", "#ffa751"]}');
  
INSERT INTO product_localizations (product_id, locale, name, description) VALUES
    ('63406c8d-67c2-43c1-a8a7-e7e79b7ac2c3', 'en', 'Frozen', 'Gradient color for your nickname'),
    ('d9607e73-ad81-46e6-b0af-ee33b5987e16', 'en', 'DeepSpace', 'Gradient color for your nickname'),
    ('2a65d174-8896-411a-a6dd-8341b3620171', 'en', 'Mango', 'Gradient color for your nickname');

INSERT INTO product_localizations (product_id, locale, name, description) VALUES
    ('63406c8d-67c2-43c1-a8a7-e7e79b7ac2c3', 'ru', 'Frozen', 'Градиентный цвет для вашего ника'),
    ('d9607e73-ad81-46e6-b0af-ee33b5987e16', 'ru', 'DeepSpace', 'Градиентный цвет для вашего ника'),
    ('2a65d174-8896-411a-a6dd-8341b3620171', 'ru', 'Mango', 'Градиентный цвет для вашего ника');