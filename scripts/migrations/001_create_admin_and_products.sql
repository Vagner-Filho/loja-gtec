-- Create admin_users table
CREATE TABLE IF NOT EXISTS admin_users (
    id SERIAL PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create products table
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    image TEXT NOT NULL,
    category TEXT NOT NULL,
    is_available BOOL NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Insert default products
INSERT INTO products (name, price, image, category, is_available) VALUES
    ('Purificador IBBL Mio Branco', 699.99, '/static/images/purificador.jpg', 'purificadores', TRUE),
    ('Bebedouro IBBL Compact', 499.99, '/static/images/bebedouro.jpg', 'bebedouros', TRUE),
    ('Válvula Redutora de Pressão 1/4', 45.99, '/static/images/peca.jpg', 'pecas', TRUE),
    ('Refil Gioviale Rpc-01 Lorenzetti', 89.99, '/static/images/refil.jpg', 'refis', TRUE);

CREATE TABLE IF NOT EXISTS services (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    image TEXT NOT NULL,
    description TEXT NOT NULL,
    is_available BOOL NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Insert default products
INSERT INTO services (name, price, image, description, is_available) VALUES
    ('Serviço de Instalação', 120.00, '/static/images/instalacao.svg', 'Instalação profissional, certificada e de garantia para produtos comprados na loja.', TRUE);

-- Create index on category for faster filtering
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);

-- Create a default admin user (username: admin, password: admin123)
-- Password hash for 'admin123' using bcrypt
INSERT INTO admin_users (username, password_hash) VALUES
    ('admin', '$2a$10$rXqK9VGKvJ5YPYz9O4vHLOGvGbkE3VJ5zJq5x5P5qWwYJ5ZJQ5ZJO');
