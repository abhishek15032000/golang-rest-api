CREATE TABLE IF NOT EXISTS users (
   id SERIAL PRIMARY KEY,
   username VARCHAR(255) NOT NULL,
   email VARCHAR(255) NOT NULL UNIQUE,
   password VARCHAR(255) NOT NULL,
   email_verified boolean default false,
   created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
   updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS blogs (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content VARCHAR(255) NOT NULL,
    user_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

create table if not exists user_profiles(
    id serial primary key,
    user_id int not null,
    profile_image varchar(255),
    created_at timestamp default current_timestamp,
    updated_at timestamp default current_timestamp,
    foreign key (user_id) references users(id) on delete cascade
);

CREATE TABLE IF NOT EXISTS otp_verification(
    id serial PRIMARY KEY,
    otp_key varchar(255) NOT NULL,
    user_id int NOT NULL,
    issued_at timestamp DEFAULT current_timestamp,
    expires_at timestamp DEFAULT current_timestamp + interval '5 minutes',
    is_used boolean DEFAULT false,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT unique_user_otp UNIQUE (user_id, otp_key)
);