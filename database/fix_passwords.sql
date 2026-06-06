USE go_replay;
UPDATE users SET password='$2a$10$26iDPYFGZCqO0wDNHIgieO5TBkbnghdoBxsrV1ekicZ0dG8Q83z4y' WHERE username='admin';
UPDATE users SET password='$2a$10$hCvCWyG.n8nVBf36P8iAOuLokOT2SKuq0Zh7m7VRvKuUvASjME9Ou' WHERE username='teacher1';
UPDATE users SET password='$2a$10$6D3ZW6PzTCK8Kwi68hjBGumgxSe4sjfaUvyxH/CyS2Hd4KO1R7jYq' WHERE username='student1';
SELECT username, password FROM users;
