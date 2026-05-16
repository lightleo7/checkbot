import os
from fastapi import FastAPI, File, Form, HTTPException, UploadFile
from pydantic import BaseModel, ValidationError
import uvicorn, time

app = FastAPI()

os.makedirs(UPLOAD_DIR, exist_ok=True)


class ImageMetadata(BaseModel):
    situation_id: int

@app.post("/new_situation")
async def new_situation(metadata: str = Form(...)):
    generated_id = int(time.time())
    os.makedirs(f"./situations/{generated_id}")

    return {"success": "success", "situation_id": generated_id}

@app.post("/upload")
async def upload_image(
    # ИСПРАВЛЕНО: принимаем metadata как чистую строку (str)
    metadata: str = Form(...), 
    image: UploadFile = File(...)
):
    # ИСПРАВЛЕНО: вручную валидируем JSON-строку через Pydantic
    try:
        parsed_metadata = ImageMetadata.model_validate_json(metadata)
    except ValidationError:
        raise HTTPException(status_code=422, detail="Invalid JSON format in metadata")

    if not image.content_type.startswith("image/"):
        raise HTTPException(status_code=400, detail="Invalid file type")

    # Получаем расширение файла
    file_extension = os.path.splitext(image.filename)[1]
    
    # Используем данные из распарсенной модели
    filename = f"situation_{parsed_metadata.situation_id}{file_extension}"
    save_path = os.path.join("situations", parsed_metadata.situation_id, filename)

    content = await image.read()
    with open(save_path, "wb") as f:
        f.write(content)

    return {"status": "success", "saved_as": filename}


if __name__ == "__main__":
    uvicorn.run(app, host="127.0.0.1", port=8000)
