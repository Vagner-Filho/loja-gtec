INSERT INTO items (name, price, image, is_available) VALUES
    ('Purificador IBBL Mio Branco', 699.99, '/static/images/purificador.jpg', TRUE),
    ('Bebedouro IBBL Compact', 499.99, '/static/images/bebedouro.jpg', TRUE),
    ('Válvula Redutora de Pressão 1/4', 45.99, '/static/images/peca.jpg', TRUE),
    ('Refil Gioviale Rpc-01 Lorenzetti', 89.99, '/static/images/refil.jpg', TRUE),
    ('Serviço de Instalação', 120.00, '/static/images/instalacao.svg', TRUE);

-- Insert default products
INSERT INTO products (category, item_id) VALUES
    ('purificadores', 1),
    ('bebedouros', 2),
    ('pecas', 3),
    ('refis', 4);

-- Insert default products
INSERT INTO services (description, item_id) VALUES
    ('Instalação profissional, certificada e de garantia para produtos comprados na loja.', 5);

-- Create index on category for faster filtering
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);

-- Create a default admin user (username: admin, password: admin123)
-- Password hash for 'admin123' using bcrypt
INSERT INTO admin_users (username, password_hash, role) VALUES
	('admin', '$2a$10$tZagoo6qCSR5NY98NVEI9.jxVF.C8ylrbXqu6lXxI5Jhu1qDwXnN.', 'admin');
