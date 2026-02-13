INSERT INTO users (name, is_admin)
VALUES 
    ('Alice', false),
    ('Bob', false),
    ('Admin', true);

INSERT INTO events (title, start_time, capacity)
VALUES 
    ('Go Workshop', '2025-02-10 18:00', 30),
    ('Music Jam', '2025-02-14 20:00', 50);


INSERT INTO bookings (event_id, user_id, status, expires_at)
VALUES (1, 1, 'confirmed', now() + INTERVAL '1 hour');


INSERT INTO bookings (event_id, user_id, status, expires_at)
VALUES (2, 2, 'booked', now() + INTERVAL '10 minutes');


INSERT INTO bookings (event_id, user_id, status, expires_at)
VALUES (1, 2, 'booked', now() - INTERVAL '1 hour');
