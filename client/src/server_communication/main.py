import requests, json, random
base_url = "http://127.0.0.1:8000"

while True:
    r = input("> ").strip()
    if r == "new":
        url = f"{base_url}/"
        response = requests.post(url, )
    # URL вашего FastAPI сервера
    url = f"{base_url}/upload"

    # Путь к картинке на вашем компьютере
    image_path = "data.png"

    with open(image_path, "rb") as f:
        files = {"image": (image_path, f, "image/png")}
        response = requests.post(url, data={"metadata": json.dumps({"situation_id": random.randint(0, 10)})}, files=files)

    # Выводим ответ сервера
    print("Статус-код:", response.status_code)
    print("Ответ сервера:", response.json())
