from client.utils import create_frame
import cv2
import numpy as np
import matplotlib.pyplot as plt
from tqdm import tqdm
from client.capture import Frame

def imshow(img):
    rgb = cv2.cvtColor(img, cv2.COLOR_BGR2RGB)

    plt.figure(figsize=(3, 3), dpi=300)
    plt.imshow(rgb)
    plt.axis("off")
    plt.show()

def get_lines_points(x1: int, y1: int, x2: int, y2: int) -> list[tuple[int, int]]:
    if y1 == y2:
        return []
    if x1 == x2:
        return []
    k = (y1 - y2)/(x1 - x2)
    b = y2 - k*x2
    points = [(y1, x1)]
    step_y = 1 if y2 > y1 else -1
    for y in range(y1, y2, step_y):
        points.append((y, int((y-b)/k)))
    points.append((y2, x2))
    return points

def create_mask(shape: tuple[int, int], lines: list[tuple[int]], area:int = 5) -> np.ndarray:
    mask = np.zeros(shape, dtype=np.uint8)
    for line in lines:
        points = get_lines_points(*line[0])
        for point in points:
            y, x = point
            left = max(x - area, 0)
            right = min(x + area, shape[1])
            mask[y, left:right] = 255
    return mask

def preprocess(frame: Frame) -> Frame:
    img = frame.img
    gray = cv2.cvtColor(img, cv2.COLOR_BGR2GRAY)
    sobelx = cv2.Sobel(gray, ddepth=cv2.CV_64F, ksize=7, dx=1, dy=0)
    old_min, old_max = sobelx.min(), sobelx.max()
    sobel_norm = np.uint8((img - old_min) * (255.0 / (old_max - old_min)))
    blur = cv2.GaussianBlur(sobel_norm, (7, 7), 0)
    edges = cv2.Canny(blur, 100, 200)
    lines = cv2.HoughLinesP(
        edges,
        rho=1,
        theta=np.pi/180,
        threshold=125,
        minLineLength=200,
        maxLineGap=20
    )
    if lines is not None:
        return create_frame(cv2.bitwise_and(img, img, mask=create_mask(img.shape[:2], lines, 15)))
    else:
        return create_frame(np.zeros(img.shape[:2], dtype=np.uint8))
