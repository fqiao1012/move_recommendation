
IF DB_ID('MovieDB') IS NULL
BEGIN
    CREATE DATABASE MovieDB;
END;
GO

USE MovieDB;
GO

IF OBJECT_ID('dbo.Movies', 'U') IS NULL
BEGIN
    CREATE TABLE dbo.Movies (
        id INT PRIMARY KEY,
        title NVARCHAR(255) NOT NULL,
        release_year INT,
        popularity FLOAT NOT NULL
    );
END;
GO

-- Reload the test data on each import.
DELETE FROM dbo.Movies;
GO

BULK INSERT dbo.Movies
FROM '/seed/movies.csv'
WITH (
    FORMAT = 'CSV',
    FIRSTROW = 2,
    FIELDQUOTE = '"',
    FIELDTERMINATOR = ',',
    ROWTERMINATOR = '0x0a',
    TABLOCK
);
GO

SELECT COUNT(*) AS total_movies
FROM dbo.Movies;
GO
