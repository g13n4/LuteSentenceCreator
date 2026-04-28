CREATE TABLE IF NOT EXISTS kanjis
(
    literal VARCHAR(1) PRIMARY KEY,
    jlpt INTEGER DEFAULT NULL,
    freq INTEGER DEFAULT NULL,
    grade INTEGER DEFAULT NULL,
    stroke_count INTEGER DEFAULT NULL,
);

CREATE TABLE IF NOT EXISTS entry_dictionaries
(
    id          SERIAL PRIMARY KEY,
    name        TEXT UNIQUE,
    description TEXT
);

CREATE TABLE IF NOT EXISTS readings
(
    id             SERIAL PRIMARY KEY,
    entry_id INT,
    reading        TEXT,
    no_kanji BOOL -- has only hiragana and katakana
);


CREATE TABLE IF NOT EXISTS sentences
(
    id             INT PRIMARY KEY,
    sentence        TEXT,
);

CREATE TABLE IF NOT EXISTS entry_dictionaries__mtm__entries
(
    id             SERIAL PRIMARY KEY,
    entry_id INT,
    ed_id INT REFERENCES entry_dictionaries (id),
);


CREATE TABLE IF NOT EXISTS readings__mtm__kanjis
(
    id  SERIAL PRIMARY KEY,
    r_id INT REFERENCES readings (id),
    literal VARCHAR(1) REFERENCES kanjis (literal)
);

CREATE TABLE IF NOT EXISTS sentences__mtm__readings
(
    r_id INT REFERENCES readings (id),
    s_id INT REFERENCES sentences (id)
);
