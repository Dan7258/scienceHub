CREATE TABLE profiles (
    id SERIAL PRIMARY KEY,
    login VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    first_Name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    middle_name VARCHAR(255),
    gender INTEGER,
    country VARCHAR(100),
    academic_degree VARCHAR(255),
    VAC VARCHAR(255),
    appointment VARCHAR(255)
);

CREATE TABLE admins (
    id SERIAL PRIMARY KEY,
    login VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL
);

INSERT INTO admins (login, password)
VALUES ('admin', '$2a$10$4MJWcLMLTCoSBm9TZhmMAugp76gUU5zEuhA6T3IAfww3ZB.U2P/mO');

CREATE TABLE publications (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    abstract TEXT NOT NULL,
    file_link VARCHAR(255) NOT NULL,
    created_at DATE NOT NULL,
    updated_at DATE NOT NULL,
    owner_id INTEGER NOT NULL REFERENCES profiles(id)
);

CREATE TABLE tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE friend_requests (
    id SERIAL PRIMARY KEY,
    requester INTEGER NOT NULL,
    receiver INTEGER NOT NULL,
    status VARCHAR(100) NOT NULL,
    created_at DATE NOT NULL
);

CREATE TABLE publication_tags (
    publications_id INTEGER REFERENCES publications(id) ON DELETE CASCADE,
    tags_id INTEGER REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (publications_id, tags_id)
);

CREATE TABLE profile_publications (
    profiles_id INTEGER REFERENCES profiles(id) ON DELETE CASCADE,
    publications_id INTEGER REFERENCES publications(id) ON DELETE CASCADE,
    PRIMARY KEY (profiles_id, publications_id)
);

CREATE TABLE subscribs (
    profiles_id INTEGER REFERENCES Profiles(id),
    subscribers_id INTEGER REFERENCES Profiles(id),
    PRIMARY KEY (profiles_id, subscribers_id)
);

