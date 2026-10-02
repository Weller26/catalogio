INSERT INTO item_statuses (user_id, name) VALUES
    (NULL, 'Planned'),
    (NULL, 'In progress'),
    (NULL, 'Completed'),
    (NULL, 'Dropped')
ON CONFLICT DO NOTHING;

INSERT INTO item_types (user_id, name) VALUES
    (NULL, 'Movie'),
    (NULL, 'Series'),
    (NULL, 'Game'),
    (NULL, 'Book')
ON CONFLICT DO NOTHING;