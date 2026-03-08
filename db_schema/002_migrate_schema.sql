UPDATE users
SET
    first_name = split_part(full_name, ' ', 1),
    last_name = split_part(full_name, ' ', 2)
WHERE full_name IS NOT NULL;