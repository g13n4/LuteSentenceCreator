CREATE TABLE IF NOT EXISTS kanjis
(
    id      SERIAL PRIMARY KEY,
    literal TEXT UNIQUE,
    jlpt INTEGER DEFAULT NULL,
    freq INTEGER DEFAULT NULL,
    grade INTEGER DEFAULT NULL,
);

CREATE TABLE IF NOT EXISTS entry_dictionary_categories
(
    id          SERIAL PRIMARY KEY,
    name        TEXT UNIQUE,
    description TEXT
);

CREATE TABLE IF NOT EXISTS entry_dictionaries
(
    id          SERIAL PRIMARY KEY,
    name        TEXT UNIQUE,
    category_id INT REFERENCES entry_dictionary_categories (id)
);

CREATE TABLE IF NOT EXISTS entries
(
    id    SERIAL PRIMARY KEY,
    entry INT UNIQUE,
);

CREATE TABLE IF NOT EXISTS readings
(
    id             SERIAL PRIMARY KEY,
    reading        TEXT,
    no_kanji BOOL -- has only hiragana and katakana
);

CREATE TABLE IF NOT EXISTS readings__mtm__kanjis
(
    r_id INT REFERENCES readings (id),
    k_id INT REFERENCES kanjis (id)
);

CREATE TABLE IF NOT EXISTS entry__mtm__readings__mtm__entry_dictionaries
(
    e_id  INT REFERENCES readings (id),
    r_id  INT REFERENCES kanjis (id),
    ed_id INT REFERENCES entries (id)
);

CREATE TABLE IF NOT EXISTS sentences
(
    id             SERIAL PRIMARY KEY,
    sentence        TEXT,
);

CREATE TABLE IF NOT EXISTS sentences__mtm__readings
(
    r_id INT REFERENCES readings (id),
    s_id INT REFERENCES sentences (id)
);
