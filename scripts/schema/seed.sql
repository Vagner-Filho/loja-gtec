INSERT INTO items (name, price, image, is_available) VALUES
    ('Serviço de Instalação', 120.00, '/static/images/instalacao.svg', TRUE),
    ('Purificador IBBL Mio Branco', 699.99, '/static/images/purificador.jpg', TRUE),
    ('Bebedouro IBBL Compact', 499.99, '/static/images/bebedouro.jpg', TRUE),
    ('Válvula Redutora de Pressão 1/4', 45.99, '/static/images/peca.jpg', TRUE),
    ('Refil Gioviale Rpc-01 Lorenzetti', 89.99, '/static/images/refil.jpg', TRUE);

-- Insert default products
INSERT INTO products (category, item_id) VALUES
    ('purificadores', 2),
    ('bebedouros', 3),
    ('pecas', 4),
    ('refis', 5);

-- Insert default products
INSERT INTO services (description, item_id) VALUES
    ('Instalação profissional, certificada e de garantia para produtos comprados na loja.', 1);

-- Create index on category for faster filtering
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
