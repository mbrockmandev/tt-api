-- add tables
CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  email TEXT UNIQUE NOT NULL,
  first_name TEXT NOT NULL,
  last_name TEXT NOT NULL,
  password TEXT NOT NULL,
  role TEXT NOT NULL CHECK (role IN ('admin', 'staff', 'user')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS books (
  id SERIAL PRIMARY KEY,
  title TEXT NOT NULL,
  author TEXT NOT NULL,
  isbn TEXT UNIQUE NOT NULL,
  published_at TIMESTAMPTZ NOT NULL,
  summary TEXT,
  thumbnail TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS genres (
  id SERIAL PRIMARY KEY,
  name TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS libraries (
  id SERIAL PRIMARY KEY,
  name TEXT UNIQUE NOT NULL,
  city TEXT NOT NULL,
  street_address TEXT NOT NULL,
  postal_code TEXT NOT NULL,
  country TEXT NOT NULL,
  phone TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS users_books (
  id SERIAL PRIMARY KEY,
  user_id INT REFERENCES users(id) ON DELETE CASCADE,
  book_id INT REFERENCES books(id) ON DELETE CASCADE,
  due_date TIMESTAMPTZ,
  returned_at TIMESTAMPTZ,
  borrowed_at TIMESTAMPTZ,
);

CREATE TABLE IF NOT EXISTS books_genres (
  id SERIAL PRIMARY KEY,
  book_id INT REFERENCES books(id) ON DELETE CASCADE,
  genre_id INT REFERENCES genres(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS users_genres (
  id SERIAL PRIMARY KEY,
  user_id INT REFERENCES users(id) ON DELETE CASCADE,
  genre_id INT REFERENCES genres(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS books_ratings (
  id SERIAL PRIMARY KEY,
  user_id INT REFERENCES users(id) ON DELETE CASCADE,
  book_id INT REFERENCES books(id) ON DELETE CASCADE,
  rating INT CHECK (
    rating BETWEEN 1
    AND 5
  )
);

CREATE TABLE IF NOT EXISTS activity_logs (
  id SERIAL PRIMARY KEY,
  user_id INT REFERENCES users(id),
  action TEXT NOT NULL CHECK (action IN ('borrow', 'return')),
  timestamp TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS books_libraries (
  id SERIAL PRIMARY KEY,
  book_id INT REFERENCES books(id) ON DELETE CASCADE,
  library_id INT REFERENCES libraries(id) ON DELETE CASCADE,
  total_copies INT NOT NULL DEFAULT 0,
  available_copies INT NOT NULL DEFAULT 0,
  borrowed_copies INT NOT NULL DEFAULT 0,
  UNIQUE(book_id, library_id)
);

CREATE TABLE IF NOT EXISTS users_libraries (
  id SERIAL PRIMARY KEY,
  user_id INT REFERENCES users(id),
  library_id INT REFERENCES libraries(id)
);

-- PlantUML notation:
-- //www.plantuml.com/plantuml/dpng/hLDFxn8n4BtlfsWugScFn2idC55ZX8GeNhmqNJf2O_-GwHJT6B-xfOdRfGn4enUMllVwvl6ONKPIICdP3ZmP6iId21Z5Zqu8enV2U1BFjk-VvoCuPUn247odVFfYC9Bqoi48MTKY9nNZju1w133OokuH586MYXPYzsxp-kDkjTdkFcScdJJB_En-ftmLmbSsPMOd8rIYOy3nQ6BlQxZKMnEFR82Od5Cu_9UeRy4Zi5adNLNvslIeqDo_KKCPgWaZ8G19fJL2ZFL7XaZAodtGatiXUSAXQex1t-OeUo3lzfmoBKViJS7wB6undK1U6cfxGw24dgqK8NkV6Ohv7_Z5a6pdTpCUBkfhiEs1D-IBu0r54_638kCF_zGj5ei2tGH-k4Qsa4FZnynjU_rRLpvvTLtjoPpp5xaEVrw5ofT2VO_9E8YatEbc9KeJOlPo9fwUycC-Vp6RhdLNFX_mlC5a7xhoUAXWXnICz-0KN8xhBMAcj0hMPDe_BsdMgKugP3kiohebU9sS_GK0
-- PNG version:
-- https://cdn-0.plantuml.com/plantuml/dpng/hLDFxn8n4BtlfsWugScFn2idC55ZX8GeNhmqNJf2O_-GwHJT6B-xfOdRfGn4enUMllVwvl6ONKPIICdP3ZmP6iId21Z5Zqu8enV2U1BFjk-VvoCuPUn247odVFfYC9Bqoi48MTKY9nNZju1w133OokuH586MYXPYzsxp-kDkjTdkFcScdJJB_En-ftmLmbSsPMOd8rIYOy3nQ6BlQxZKMnEFR82Od5Cu_9UeRy4Zi5adNLNvslIeqDo_KKCPgWaZ8G19fJL2ZFL7XaZAodtGatiXUSAXQex1t-OeUo3lzfmoBKViJS7wB6undK1U6cfxGw24dgqK8NkV6Ohv7_Z5a6pdTpCUBkfhiEs1D-IBu0r54_638kCF_zGj5ei2tGH-k4Qsa4FZnynjU_rRLpvvTLtjoPpp5xaEVrw5ofT2VO_9E8YatEbc9KeJOlPo9fwUycC-Vp6RhdLNFX_mlC5a7xhoUAXWXnICz-0KN8xhBMAcj0hMPDe_BsdMgKugP3kiohebU9sS_GK0
