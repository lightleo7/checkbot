import cv2
import numpy as np
from client.capture import Frame
from client.capture.data_classes import LineSegment, FrameWithLines

def create_mask(shape: tuple[int, int], lines: list[LineSegment], area:int = 5) -> np.ndarray:
    mask = np.zeros(shape, dtype=np.uint8)
    for line in lines:
        for (x, y) in line.points():
            left = max(x - area, 0)
            right = min(x + area, shape[1])
            mask[y, left:right] = 255
    return mask

def detect_rails(frame: Frame) -> Frame:
    gray = cv2.cvtColor(frame.img, cv2.COLOR_BGR2GRAY)
    sobelx = cv2.Sobel(gray, ddepth=cv2.CV_64F, ksize=7, dx=1, dy=0)
    old_min, old_max = sobelx.min(), sobelx.max()
    sobel_norm = np.uint8((sobelx - old_min) * (255.0 / (old_max - old_min)))
    blur = cv2.GaussianBlur(sobel_norm, (7, 7), 0)
    edges = cv2.Canny(blur, 100, 200)
    lines = [LineSegment(line) for line in cv2.HoughLinesP(edges, rho=1, theta=np.pi/180, threshold=125, minLineLength=200, maxLineGap=20)]
    if lines is not None:
        lines = sorted(lines, lambda line: line.square_of_length, reverse=True)[:2]
        frame.update(cv2.bitwise_and(frame.img, frame.img, mask=create_mask(frame.img.shape[:2], lines, 15)))
        return FrameWithLines(frame, lines)
    else:
        return FrameWithLines(frame.update(np.zeros(frame.img.shape[:2], dtype=np.uint8)), [])
