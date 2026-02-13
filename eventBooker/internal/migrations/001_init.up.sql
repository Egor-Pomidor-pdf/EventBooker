create table events(
    id serial primary key,
    title text not null,
    start_time timestamp not null,
    capacity int not null check (capacity >0),
    created_at timestamp default now()
);

create table users(
    id serial primary key,
    name varchar(30) not null,
    is_admin boolean default false
);

create table bookings (
    id serial primary key,
    event_id int not null references events(id) on delete cascade,
    user_id int not null references users(id) on delete cascade,
    status varchar(20) not null check (status in ('booked', 'confirmed', 'cancelled')),
    created_at timestamp default now(),
    expires_at timestamp not null, 
    unique(user_id, event_id)
);