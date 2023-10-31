-- Create a table to store events
CREATE TABLE events (
    event_id SERIAL PRIMARY KEY,
    event_type VARCHAR(255) NOT NULL,
    event_title VARCHAR(255),
    user_id INT NOT NULL, -- Assuming a reference to the user table
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL
);

-- Create an index for user_id for faster user-specific event retrieval
CREATE INDEX idx_user_id ON events (user_id);

CREATE USER bookings WITH PASSWORD 'bookings';
GRANT SELECT, INSERT, UPDATE, DELETE ON events TO bookings;
GRANT USAGE, SELECT ON SEQUENCE events_event_id_seq TO bookings;