INSERT INTO items (name, price, is_available) VALUES
    ('Serviço de Instalação', 120.00, TRUE),
    ('Purificador IBBL Mio Branco', 699.99, TRUE),
    ('Bebedouro IBBL Compact', 499.99, TRUE),
    ('Válvula Redutora de Pressão 1/4', 45.99, TRUE),
    ('Refil Gioviale Rpc-01 Lorenzetti', 89.99, TRUE);

INSERT INTO products (category, item_id, description, sku) VALUES
    ('purificadores', 2, 'Purificador de água IBBL Mio Branco 127V', 'IBB-MIO-BRC-127'),
    ('bebedouros', 3, 'Bebedouro IBBL Compact', 'IBB-COMP-001'),
    ('pecas', 4, 'Válvula Redutora de Pressão 1/4', 'VRP-14-001'),
    ('refis', 5, 'Refil Gioviale RPC-01 Lorenzetti', 'LOR-RPC-01');

INSERT INTO services (description, item_id) VALUES
    ('Instalação profissional, certificada e de garantia para produtos comprados na loja.', 1);

INSERT INTO brands (name) VALUES
    ('IBBL'),
    ('Lorenzetti');

INSERT INTO product_dimensions (product_id, weight, length, width, height) VALUES
    (1, 2.500, 30.00, 20.00, 40.00),
    (2, 5.000, 35.00, 25.00, 45.00),
    (3, 0.150, 5.00, 5.00, 5.00),
    (4, 0.300, 10.00, 8.00, 15.00);

INSERT INTO product_images (product_id, image_url, display_order, is_primary) VALUES
    (1, '/static/images/purificador.jpg', 0, TRUE),
    (2, '/static/images/bebedouro.jpg', 0, TRUE),
    (3, '/static/images/peca.jpg', 0, TRUE),
    (4, '/static/images/refil.jpg', 0, TRUE);

INSERT INTO product_technical_specs (product_id, spec_key, spec_value, display_order) VALUES
    (1, 'Voltagem', '127V', 1),
    (1, 'Capacidade', '10L/h', 2),
    (1, 'Temperatura', 'Morna/Fria', 3),
    (1, 'Garantia', '12 meses', 4),
    (2, 'Voltagem', '127V', 1),
    (2, 'Capacidade', '15L/h', 2),
    (2, 'Temperatura', 'Fria', 3),
    (2, 'Garantia', '12 meses', 4),
    (3, 'Pressão Máx', '10 kgf/cm²', 1),
    (3, 'Pressão Mín', '1 kgf/cm²', 2),
    (3, 'Conexão', '1/4 polegada', 3),
    (4, 'Compatibilidade', 'Gioviale', 1),
    (4, 'Vida Útil', '6 meses', 2),
    (4, 'Fluxo', '60 L/h', 3);

INSERT INTO offers (product_id, offer_price, start_date, end_date, is_active) VALUES
    (1, 599.99, '2026-01-01', '2026-12-31', TRUE),
    (4, 79.99, '2026-01-01', '2026-06-30', TRUE);

INSERT INTO banners (image_path, title, link_url, display_order, is_active) VALUES
    ('/static/images/banner-oferta-purificador.jpg', 'Oferta especial: Purificador IBBL Mio', '/?category=purificadores', 1, TRUE);

CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
