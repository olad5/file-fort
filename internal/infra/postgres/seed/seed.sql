INSERT INTO users (id, first_name, last_name, email, role, password) VALUES
	('5cf37266-3473-4006-984f-9325122678b7', 'jason','ford', 'admin@app.com', 'admin', '$2a$04$X.3Soh0RZYTMyzMxlpOjOeFy8qLSFLtAGc555Dkwys2ZgyGvscslC')
ON CONFLICT DO NOTHING;
