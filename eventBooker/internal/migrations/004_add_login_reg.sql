alter table users add column email varchar(100) unique not null;
alter table users add column password_hash text not null;