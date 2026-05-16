import os
from fastapi import FastAPI, File, Form, HTTPException, UploadFile
from pydantic import BaseModel, Json

app = FastAPI()

UPLOAD_DIR = "./uploaded_images"
os.makedirs(UPLOAD_DIR, exist_ok=True)


class ImageMetadata(BaseModel):
    user_id: int
    category: str


@app.post("/upload")
async def upload_image(
    metadata: Json[ImageMetadata] = Form(...), image: UploadFile = File(...)
):
    if not image.content_type.startswith("image/"):
        raise HTTPException(status_code=400, detail="Invalid file type")

    filename = f"user_{metadata.user_id}_{metadata.category}{os.path.splitext(image.filename)[1]}"
    save_path = os.path.join(UPLOAD_DIR, filename)

    content = await image.read()
    with open(save_path, "wb") as f:
        f.write(content)

    return {
        "status": "success",
        "filename": filename,
        "received_metadata": metadata.dict(),
    }


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="127.0.0.1", port=8000)

