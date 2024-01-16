package models

import (
    "database/sql"
    "log"
)

type Image struct {
    ImageID int
    UserID  int
    ImageURL string
    Description string
    UploadTimestamp string
}

type ImageRepository struct {
    DB *sql.DB
}

func (repo *ImageRepository) GetAllImages() ([]Image, error) {
    rows, err := repo.DB.Query(`SELECT * FROM Public."Images" WHERE "Images"."UserID" = 1`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    images := []Image {}
    defer rows.Close()
    for rows.Next() {
        var img Image
        err := rows.Scan(&img.ImageID, &img.UserID, &img.ImageURL, &img.Description, &img.UploadTimestamp)
        if err != nil {
            log.Fatal(err)
        }
        images = append(images, img)
    }
    return images, nil
}

func (i *Image) CreateImage() {

}

func (i *Image) DeleteImage() {

}
