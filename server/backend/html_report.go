package main

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
)

type DamageReport struct {
	ID           string         // Уникальный ID для имени файла
	Date         string         // Дата
	TimeSpotted  string         // Время когда увидел
	TimeReceived string         // Время принятия на сервер
	TypeName     string         // Тип
	Coordinates  string         // Координаты
	ImagePaths   []string       // Пути к файлам картинок (слайсы строк)
	ImagesBase64 []template.URL // Закодированные картинки для HTML
}

const lightReportTemplate = `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Отчет #{{.ID}}</title>
    <style>
        :root {
            --bg-main: #f8fafc;
            --bg-card: #ffffff;
            --text-main: #0f172a;
            --text-muted: #64748b;
            --border: #e2e8f0;
            --accent: #0284c7;
        }
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; 
            background-color: var(--bg-main); 
            color: var(--text-main);
            padding: 40px;
            min-height: 100vh;
        }
        .header {
            margin-bottom: 32px;
            border-bottom: 1px solid var(--border);
            padding-bottom: 20px;
        }
        .header h1 { font-size: 32px; font-weight: 700; letter-spacing: -0.025em; color: var(--text-main); }
        .header h1 span { color: var(--accent); }
        
        .layout {
            display: grid;
            grid-template-columns: 1fr;
            gap: 30px;
        }
        @media (min-width: 1024px) {
            .layout { grid-template-columns: 1fr 1fr; }
        }

        .metrics-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
            gap: 20px;
            align-content: start;
        }
        .metric-card {
            background-color: var(--bg-card);
            border: 1px solid var(--border);
            border-radius: 12px;
            padding: 24px;
            box-shadow: 0 1px 3px rgba(0,0,0,0.02);
            transition: box-shadow 0.2s;
        }
        .metric-card:hover { box-shadow: 0 4px 12px rgba(0,0,0,0.05); }
        .metric-label {
            font-size: 13px;
            text-transform: uppercase;
            letter-spacing: 0.05em;
            color: var(--text-muted);
            margin-bottom: 8px;
            font-weight: 600;
        }
        .metric-value { font-size: 20px; font-weight: 500; color: var(--text-main); }
        
        .full-width { grid-column: 1 / -1; }

        /* Галерея изображений */
        .gallery-container {
            display: flex;
            flex-direction: column;
            gap: 20px;
        }
        .image-box {
            background-color: var(--bg-card);
            border: 1px solid var(--border);
            border-radius: 12px;
            overflow: hidden;
            box-shadow: 0 1px 3px rgba(0,0,0,0.02);
        }
        .image-box img {
            width: 100%;
            height: auto;
            max-height: 500px;
            object-fit: contain;
            display: block;
            background: #f1f5f9;
        }
        .no-image { 
            background-color: var(--bg-card);
            border: 1px solid var(--border);
            border-radius: 12px;
            padding: 60px;
            text-align: center;
            color: var(--text-muted); 
            font-size: 16px; 
        }
    </style>
</head>
<body>

    <div class="header">
        <h1>Отчет по инциденту: <span>#{{.ID}}</span></h1>
    </div>

    <div class="layout">
        <div class="metrics-grid">
            <div class="metric-card">
                <div class="metric-label">Дата</div>
                <div class="metric-value">{{.Date}}</div>
            </div>
            <div class="metric-card">
                <div class="metric-label">Время обнаружения</div>
                <div class="metric-value">{{.TimeSpotted}}</div>
            </div>
            <div class="metric-card full-width">
                <div class="metric-label">Время получения</div>
                <div class="metric-value">{{.TimeReceived}}</div>
            </div>
            <div class="metric-card full-width">
                <div class="metric-label">Тип</div>
                <div class="metric-value">{{.TypeName}}</div>
            </div>
            <div class="metric-card full-width">
                <div class="metric-label">Координаты</div>
                <div class="metric-value" style="font-family: monospace; color: var(--accent); font-weight: 600;">{{.Coordinates}}</div>
            </div>
        </div>

        <div class="gallery-container">
            {{if .ImagesBase64}}
                {{range .ImagesBase64}}
                    <div class="image-box">
                        <img src="{{.}}" alt="Фото повреждения">
                    </div>
                {{end}}
            {{else}}
                <div class="no-image">Изображения отсутствуют</div>
            {{end}}
        </div>
    </div>

</body>
</html>`

func imageToBase64(path string) template.URL {
	if path == "" {
		return ""
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	encoded := base64.StdEncoding.EncodeToString(bytes)
	return template.URL(fmt.Sprintf("data:image/jpeg;base64,%s", encoded))
}

func ExportSingleReport(report DamageReport, outputDir string) error {
	for _, path := range report.ImagePaths {
		if encoded := imageToBase64(path); encoded != "" {
			report.ImagesBase64 = append(report.ImagesBase64, encoded)
		}
	}

	tmpl, err := template.New("light_report").Parse(lightReportTemplate)
	if err != nil {
		return fmt.Errorf("ошибка шаблона: %w", err)
	}

	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return err
	}

	filename := filepath.Join(outputDir, fmt.Sprintf("report_%s.html", report.ID))
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("ошибка создания файла: %w", err)
	}
	defer file.Close()

	return tmpl.Execute(file, report)
}

// func main() {
// 	report := DamageReport{
// 		ID:           "INC-2026-A",
// 		Date:         "01.06.2026",
// 		TimeSpotted:  "19:15:00",
// 		TimeReceived: "19:15:32",
// 		TypeName:     "Деформация внешнего контура здания",
// 		Coordinates:  "55.755823, 37.617298",
// 		ImagePaths: []string{
// 			"img_general.jpg", // Общий план
// 			"img_close_up.jpg", // Крупный план повреждения
// 		},
// 	}

// 	err := ExportSingleReport(report, "./reports_light")
// 	if err != nil {
// 		fmt.Printf("Ошибка: %v\n", err)
// 		return
// 	}

// 	fmt.Println("Светлый отчет успешно сгенерирован по отдельности в папку ./reports_light")
// }
