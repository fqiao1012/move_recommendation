
USE MovieDB;
GO

-- Test 1: Count movies
SELECT COUNT(*) AS total_movies
FROM dbo.Movies;
GO

-- Test 2: Prefix search
SELECT TOP (10)
    id,
    title,
    popularity
FROM dbo.Movies
WHERE title LIKE N'Star%'
ORDER BY popularity DESC;
GO

-- Test 3: Retrieve most popular movies
SELECT TOP (5)
    id,
    title,
    popularity
FROM dbo.Movies
ORDER BY popularity DESC;
GO
