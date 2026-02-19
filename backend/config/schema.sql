-- 创建 products 表
CREATE TABLE IF NOT EXISTS products (
    id BIGSERIAL PRIMARY KEY,
    product_code VARCHAR(255) NOT NULL UNIQUE,
    ingredient_code VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    amount DECIMAL(8,2) NOT NULL,
    unit VARCHAR(20) NOT NULL,
    description TEXT,
    price DECIMAL(8,2) NOT NULL,
    allergen_type VARCHAR(20) DEFAULT 'none',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    -- deleted_at TIMESTAMP NULL
);

-- 为 products 表创建索引（PostgreSQL 语法）
CREATE INDEX IF NOT EXISTS idx_products_ingredient_code ON products (ingredient_code);
-- CREATE INDEX IF NOT EXISTS idx_products_deleted_at ON products (deleted_at);

-- 创建 ingredients 表
CREATE TABLE IF NOT EXISTS ingredients (
    id BIGSERIAL PRIMARY KEY,
    ingredient_code VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    -- deleted_at TIMESTAMP NULL
);

-- 为 ingredients 表创建索引（PostgreSQL 语法）
-- CREATE INDEX IF NOT EXISTS idx_ingredients_deleted_at ON ingredients (deleted_at);

-- 创建 dishes 表
CREATE TABLE IF NOT EXISTS dishes (
    id BIGSERIAL PRIMARY KEY,
    dish_code VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    -- 菜品的具体做法
    recipe TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    -- deleted_at TIMESTAMP NULL
);

-- 为 dishes 表创建索引（PostgreSQL 语法）
-- CREATE INDEX IF NOT EXISTS idx_dishes_deleted_at ON dishes (deleted_at);


CREATE TABLE IF NOT EXISTS dish_ingredients (
    id BIGSERIAL PRIMARY KEY,
    dish_code VARCHAR(255) NOT NULL,
    ingredient_code VARCHAR(255) NOT NULL,
    quantity VARCHAR(100), -- 用量：如"200g"、"适量"、"2个"
    note TEXT, -- 对该食材的烹饪说明
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    -- deleted_at TIMESTAMP NULL
);

CREATE INDEX IF NOT EXISTS idx_dish_ingredients_ingredient_code ON dish_ingredients (ingredient_code);
CREATE UNIQUE INDEX IF NOT EXISTS idx_dish_ingredients_dish_ingredient_code ON dish_ingredients (dish_code, ingredient_code);
-- CREATE INDEX IF NOT EXISTS idx_dish_ingredients_deleted_at ON dish_ingredients (deleted_at);

CREATE TABLE IF NOT EXISTS dish_images (
    id BIGSERIAL PRIMARY KEY,
    dish_code VARCHAR(255) NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    image_url VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    -- deleted_at TIMESTAMP NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_dish_images_dish_code_sort_order ON dish_images (dish_code, sort_order);
CREATE INDEX IF NOT EXISTS idx_dish_images_sort_order ON dish_images (sort_order);
-- CREATE INDEX IF NOT EXISTS idx_dish_images_deleted_at ON dish_images (deleted_at);

CREATE TABLE IF NOT EXISTS ingredient_images (
    id BIGSERIAL PRIMARY KEY,
    ingredient_code VARCHAR(255) NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    image_url VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    -- deleted_at TIMESTAMP NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_ingredient_images_ingredient_code_sort_order ON ingredient_images (ingredient_code, sort_order);
CREATE INDEX IF NOT EXISTS idx_ingredient_images_sort_order ON ingredient_images (sort_order);
-- CREATE INDEX IF NOT EXISTS idx_ingredient_images_deleted_at ON ingredient_images (deleted_at);


CREATE TABLE IF NOT EXISTS product_images (
    id BIGSERIAL PRIMARY KEY,
    product_code VARCHAR(255) NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    image_url VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    -- deleted_at TIMESTAMP NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_product_images_product_code_sort_order ON product_images (product_code, sort_order);
CREATE INDEX IF NOT EXISTS idx_product_images_sort_order ON product_images (sort_order);
-- CREATE INDEX IF NOT EXISTS idx_product_images_deleted_at ON product_images (deleted_at);
