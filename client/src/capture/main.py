from pathlib import Path
import cv2

ROOT_DATA_PATH = Path("data/train/").resolve()

vc = cv2.VideoCapture(ROOT_DATA_PATH / "rails1.mp4")

while True:
    success, frame = vc.read()
    if not success:
        break
    cv2.imshow("video", frame)
    cv2.waitKey(1)
cv2.destroyAllWindows()

def get1():
    return 2
