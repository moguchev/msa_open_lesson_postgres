# Кейс

Вы — DBA/разработчик в сервисе доставки. Есть таблица событий `courier_events`: статусы заказов, гео-обновления курьеров, ошибки приложений. Поток — десятки тысяч событий в минуту. Нужно:

1. ускорить выборки «за последние N дней/часов»,
2. упростить ретеншн (удаление старых данных),
3. не потерять редкие «запаздывающие» события.

## Исходные требования

* Храним 9 месяцев данных, всё старше — удаляем «атомарно».
* Частые запросы: последние 7/30 дней, а также по конкретному `courier_id` и/или `order_id`.
* Бывают запаздывающие события на 3–10 дней.
* Нужна уникальность `(event_id)` в пределах всей таблицы.
* Поддержать быструю массовую загрузку (батчи).

---

## Задание (этапы)

### 1) Спроектировать схему и выбрать стратегию партиционирования

* Предложите стратегию: **RANGE по времени** (например, по `event_time`) + **сабпартиционирование LIST/HASH** по `courier_id` или `order_id` — аргументируйте выбор.
* Объясните политику размера и шага партиций (дни/недели/месяцы) с привязкой к частоте запросов.
* Продумайте **DEFAULT-партицию** для «запаздывающих» событий и план регулярного «переселения» их в целевые партиции.

### 2) Реализовать схему (DDL)

1. Создайте родительскую таблицу и партиции.
2. Обеспечьте глобальную уникальность событий:

   * через составной первичный ключ `(event_id, event_time)` **или** уникальный индекс, совместимый с партиционированием (подумайте, почему просто `PRIMARY KEY (event_id)` не взлетит).
3. Создайте **локальные индексы** на партициях под типовые запросы:

   * по времени,
   * по `(courier_id, event_time)`,
   * опционально `(order_id)`.
4. Подготовьте **RETENTION-механику**: быстрое удаление старых партиций и автоматическое создание новых.

### 3) Сгенерировать и загрузить тестовые данные

* Сгенерируйте \~5–20 млн строк (объём на ваше усмотрение, главное — чтобы разница была видна).
* Смоделируйте распределение: 80% последних 7 дней, 20% — «хвост» за 9 месяцев, и \~1% «запаздывающих» (event\_time в прошлом, но `ingested_at` — сегодня).
* Загрузку сделайте батчами (COPY/UNLOGGED staging → INSERT with ON CONFLICT/partition routing).

### 4) Замерить до/после (бенчмарки)

Для **до** (без партиций) и **после** (с партициями) замерьте:

* `EXPLAIN (ANALYZE, BUFFERS)` для запросов:

  1. последние 7 дней по всем событиям,
  2. последние 30 дней по `courier_id`,
  3. конкретный `order_id` за последние 14 дней,
  4. агрегация: количество событий на курьера по дням за 30 дней,
  5. массовое удаление старше 9 месяцев (до — `DELETE`, после — `DROP PARTITION`).
* Сравните время, количество прочитанных страниц и фактическое **partition pruning** (убедитесь, что сканируются только нужные партиции).

### 5) Обработать «запаздывающие» события

* Реализуйте DEFAULT-партицию и ежедневный job/скрипт, который:

  1. находит записи из DEFAULT, которым уже «нашлась» целевая партиция,
  2. переносит их (например, `INSERT ... ON CONFLICT DO NOTHING; DELETE` из DEFAULT),
  3. логирует объёмы.
* Покажите, как избежать долгих блокировок (батчи/лимиты/индексы).

### 6) Ретеншн и обслуживание

* Напишите SQL/PLpgSQL-скрипт: создать партиции на следующий месяц/неделю; удалить партиции старше 9 месяцев.
* Объясните, как этот скрипт повесить на `pg_cron`/`cron`.
* Покажите, что `DROP TABLE ...` для партиции мгновенно освобождает данные (и почему это быстрее `DELETE`).

### 7) (Опционально) Partition-wise операции

* Продемонстрируйте partition-wise aggregate или join (например, join с партиционированной `orders` по `order_id`/дате) и покажите, что PostgreSQL делает это по частям.

---

## Шаблоны кода (можно копировать и адаптировать)

### Родитель и партиции (пример: RANGE месяц + HASH по courier\_id на 4 сабпартиции)

```sql
CREATE SCHEMA IF NOT EXISTS logistics;

CREATE TABLE logistics.courier_events (
  event_id        BIGINT       NOT NULL,
  event_time      TIMESTAMPTZ  NOT NULL,
  ingested_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
  courier_id      BIGINT       NOT NULL,
  order_id        BIGINT,
  event_type      TEXT         NOT NULL, -- e.g. location_update, status_change, error
  payload         JSONB        NOT NULL,
  -- глобальная уникальность: составной ключ, включающий ключ партиционирования
  CONSTRAINT courier_events_pk PRIMARY KEY (event_id, event_time)
) PARTITION BY RANGE (event_time);

-- Помесячные партиции за последние 9 месяцев + текущий + следующий
-- Пример для одного месяца:
CREATE TABLE logistics.courier_events_2025_08
  PARTITION OF logistics.courier_events
  FOR VALUES FROM ('2025-08-01') TO ('2025-09-01')
  PARTITION BY HASH (courier_id);

-- Сабпартиции по courier_id (4 корзины)
DO $$
DECLARE i int;
BEGIN
  FOR i IN 0..3 LOOP
    EXECUTE format($f$
      CREATE TABLE logistics.courier_events_2025_08_h%s
      PARTITION OF logistics.courier_events_2025_08
      FOR VALUES WITH (MODULUS 4, REMAINDER %s)$f$, i, i);
  END LOOP;
END$$;

-- DEFAULT для «запаздывающих»
CREATE TABLE logistics.courier_events_default
  PARTITION OF logistics.courier_events DEFAULT;
```

### Индексы (локальные)

```sql
-- На всех сабпартициях (пример для одной; автоматизируйте через DO-блок)
CREATE INDEX ON logistics.courier_events_2025_08_h0 (event_time);
CREATE INDEX ON logistics.courier_events_2025_08_h0 (courier_id, event_time);
CREATE INDEX ON logistics.courier_events_2025_08_h0 (order_id);
```

### Генерация данных

```sql
-- Быстрый генератор в CTE-батчах
INSERT INTO logistics.courier_events (event_id, event_time, ingested_at, courier_id, order_id, event_type, payload)
SELECT
  gs AS event_id,
  now() - (random()*interval '270 days') AS event_time, -- 9 мес разброс
  now() AS ingested_at,
  (100000 + (random()*5000)::int) AS courier_id,
  (200000 + (random()*100000)::int) AS order_id,
  (ARRAY['location_update','status_change','error'])[1 + (random()*2)::int] AS event_type,
  jsonb_build_object('lat', 55 + random(), 'lon', 37 + random(), 'battery', (random()*100)::int)
FROM generate_series(1, 5000000) gs;

-- Добавьте «запаздывающие» (event_time в прошлом, но вставка сегодня упадёт в DEFAULT, если нет целевой партиции)
```

### Типовые запросы для бенчмарка

```sql
EXPLAIN (ANALYZE, BUFFERS)
SELECT *
FROM logistics.courier_events
WHERE event_time >= now() - interval '7 days';

EXPLAIN (ANALYZE, BUFFERS)
SELECT *
FROM logistics.courier_events
WHERE event_time >= now() - interval '30 days'
  AND courier_id = 101234;

EXPLAIN (ANALYZE, BUFFERS)
SELECT *
FROM logistics.courier_events
WHERE order_id = 250123
  AND event_time >= now() - interval '14 days';

EXPLAIN (ANALYZE, BUFFERS)
SELECT courier_id, date_trunc('day', event_time) d, count(*)
FROM logistics.courier_events
WHERE event_time >= now() - interval '30 days'
GROUP BY courier_id, d
ORDER BY d DESC, courier_id;
```

### Ретеншн/обслуживание (скетч)

```sql
-- Создание партиции на следующий месяц
DO $$
DECLARE
  start_ts date := date_trunc('month', now() + interval '1 month')::date;
  end_ts   date := (start_ts + interval '1 month')::date;
  m text := to_char(start_ts, 'YYYY_MM');
  i int;
BEGIN
  EXECUTE format('CREATE TABLE IF NOT EXISTS logistics.courier_events_%s PARTITION OF logistics.courier_events FOR VALUES FROM (%L) TO (%L) PARTITION BY HASH (courier_id)', m, start_ts, end_ts);
  FOR i IN 0..3 LOOP
    EXECUTE format('CREATE TABLE IF NOT EXISTS logistics.courier_events_%s_h%s PARTITION OF logistics.courier_events_%s FOR VALUES WITH (MODULUS 4, REMAINDER %s)', m, i, m, i);
    EXECUTE format('CREATE INDEX IF NOT EXISTS courier_events_%s_h%s_event_time_idx ON logistics.courier_events_%s_h%s (event_time)', m, i, m, i);
    EXECUTE format('CREATE INDEX IF NOT EXISTS courier_events_%s_h%s_courier_time_idx ON logistics.courier_events_%s_h%s (courier_id, event_time)', m, i, m, i);
  END LOOP;
END$$;

-- Удаление партиций старше 9 месяцев
DO $$
DECLARE
  cutoff date := date_trunc('month', now() - interval '9 months')::date;
  r record;
BEGIN
  FOR r IN
    SELECT relname
    FROM pg_class c
    JOIN pg_namespace n ON n.oid = c.relnamespace
    WHERE n.nspname = 'logistics'
      AND relname ~ '^courier_events_\d{4}_\d{2}$'
  LOOP
    IF to_date(substring(r.relname from '(\d{4}_\d{2})'), 'YYYY_MM') < cutoff THEN
      EXECUTE format('DROP TABLE IF EXISTS logistics.%I CASCADE', r.relname);
    END IF;
  END LOOP;
END$$;
```

### Переселение из DEFAULT (ежедневно)

```sql
-- Пример: переносим по 100k строк за проход
WITH moved AS (
  INSERT INTO logistics.courier_events (event_id, event_time, ingested_at, courier_id, order_id, event_type, payload)
  SELECT event_id, event_time, ingested_at, courier_id, order_id, event_type, payload
  FROM logistics.courier_events_default
  WHERE event_time >= date_trunc('month', now() - interval '9 months') -- в горизонте хранения
  ORDER BY event_time
  LIMIT 100000
  ON CONFLICT DO NOTHING
  RETURNING event_id, event_time
)
DELETE FROM logistics.courier_events_default d
USING moved m
WHERE d.event_id = m.event_id AND d.event_time = m.event_time;
```

---

## Что сдать

1. Краткую записку (1–2 стр.): выбранная стратегия, почему такой шаг партиций, как решены «запаздывающие», индексация, ретеншн.
2. DDL/скрипты (создание/индексы/cron-обслуживание/переселение DEFAULT).
3. Скрипт генерации/загрузки данных.
4. `EXPLAIN (ANALYZE, BUFFERS)` «до» и «после» с сравнением (таблица: время, shared hit/read, кол-во отсканированных партиций).
5. Короткие выводы: где выиграли, где узкое место, что бы поменяли при x10 объёме.

---

## Подсказки/грабли (если застрянете)

* Уникальные и первичные ключи в партиционированных таблицах должны включать **партиционный ключ**.
* **Глобальных** уникальных индексов нет: только локальные (Postgres 17).
* Следите, чтобы **CHECK-ограничения** партиций были точными — иначе pruning не сработает.
* Непродуманный размер партиции = либо слишком много файлов (overhead), либо слабое pruning.
* «Запаздывающие» без DEFAULT → ошибка вставки. С DEFAULT без ротации → пухнет партиция.
* COPY в staging-таблицу (UNLOGGED) + вставка «кусками» часто быстрее прямой вставки.
