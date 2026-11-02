# Catalogio

Веб-приложение для ведения личного каталога (фильмы, сериалы, книги, игры) с JWT-аутентификацией, сбором метрик в VictoriaMetrics, агрегацией логов через Vector в VictoriaLogs, дашбордами в Grafana и алертингом в Telegram через vmalert и Alertmanager.

**Стек:** Go 1.27 (`net/http`), PostgreSQL 18, VictoriaMetrics, Vector, VictoriaLogs, Grafana, vmalert, Alertmanager, Docker, Kubernetes.

---

## Развёртывание в Kubernetes

### 1. Переменные окружения
Создайте файл `.env` из примера и укажите секреты (пароли, JWT, токен Telegram-бота и chat ID):
```bash
cp .env.example .env
```

### 2. Установка VictoriaMetrics (Helm)
```bash
helm repo add vm https://victoriametrics.github.io/helm-charts/
helm repo update
helm install vm-single vm/victoria-metrics-single -n monitoring --create-namespace -f helm/guide-vmsingle-values.yaml
```

### 3. Запуск всех компонентов
Скрипт создаст неймспейсы, подставит секреты и применит все манифесты:
```bash
./scripts/apply-manifests.sh
```

### 4. Проброс портов
```bash
./scripts/port-forward.sh
```

### Доступ к сервисам:
- **Catalogio (приложение):** http://localhost:8080
- **Grafana:** http://localhost:3000 (логин: `admin`, пароль из `.env`)
- **VictoriaLogs:** http://localhost:9428
- **Alertmanager:** http://localhost:9093
- **Vmalert:** http://localhost:8880
- **VictoriaMetrics:** http://localhost:8428