CREATE TABLE IF NOT EXISTS kanjis
(
    id INT PRIMARY KEY,
    literal VARCHAR(1),
    jlpt INTEGER DEFAULT NULL,
    freq INTEGER DEFAULT NULL,
    grade INTEGER DEFAULT NULL,
    stroke_count INTEGER DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS dictionary_categories
(
    id          SERIAL PRIMARY KEY,
    name        TEXT UNIQUE,
    description TEXT
);

CREATE TABLE IF NOT EXISTS dictionaries
(
    id          SERIAL PRIMARY KEY,
    name        TEXT UNIQUE,
    category    INTEGER,
    number INTEGER
);

CREATE TABLE IF NOT EXISTS readings
(
    id    INT PRIMARY KEY,
    entry INT,
    reading        TEXT,
    kanji BOOLEAN, -- has only hiragana and katakana
    in_news BOOLEAN
);


CREATE TABLE IF NOT EXISTS sentences
(
    id             INT PRIMARY KEY,
    sentence        TEXT,
    isFiltered BOOLEAN DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS dictionaries__mtm__entries
(
    entry INT,
    d_id INT REFERENCES dictionaries (id),
    dc_id INT REFERENCES dictionary_categories (id),
    CONSTRAINT dictionaries__mtm__entries_pkey PRIMARY KEY (entry, d_id, dc_id)
    );


CREATE TABLE IF NOT EXISTS readings__mtm__kanjis
(
    r_id INT REFERENCES readings (id),
    k_id INT REFERENCES kanjis (id),
    CONSTRAINT readings__mtm__kanjis_pkey PRIMARY KEY (r_id, k_id)
);

CREATE TABLE IF NOT EXISTS sentences__mtm__readings
(
    r_id INT REFERENCES readings (id),
    s_id INT REFERENCES sentences (id),
    CONSTRAINT sentences__mtm__readings_pkey PRIMARY KEY (r_id, s_id)
);

CREATE TABLE IF NOT EXISTS db_state
(
    id INT,
    status INT
);

