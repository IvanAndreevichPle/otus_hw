# Документация по API календаря

## REST (через grpc-gateway)

- POST   `/v1/events` — создать событие
- PUT    `/v1/events/{id}` — обновить событие
- DELETE `/v1/events/{id}` — удалить событие
- GET    `/v1/events/day` — события за день (userId, periodStart, periodEnd)
- GET    `/v1/events/week` — события за неделю (userId, periodStart, periodEnd)
- GET    `/v1/events/month` — события за месяц (userId, periodStart, periodEnd)

### Пример структуры события (JSON)
```json
{
  "id": "b3b1c2e0-1234-4a5b-8c2d-1e2f3a4b5c6d",
  "title": "Test Event",
  "startTime": "2024-07-19T10:00:00Z",
  "durationSeconds": 3600,
  "description": "Описание события",
  "userId": "user1",
  "notifyBeforeMinutes": 10
}
```

## gRPC (EventService)

- CreateEvent(CreateEventRequest) returns (CreateEventResponse)
- UpdateEvent(UpdateEventRequest) returns (UpdateEventResponse)
- DeleteEvent(DeleteEventRequest) returns (DeleteEventResponse)
- ListEventsForDay(ListEventsRequest) returns (ListEventsResponse)
- ListEventsForWeek(ListEventsRequest) returns (ListEventsResponse)
- ListEventsForMonth(ListEventsRequest) returns (ListEventsResponse)

### Пример структуры Event (protobuf)
```proto
message Event {
  string id = 1;
  string title = 2;
  string start_time = 3;
  int64 duration_seconds = 4;
  string description = 5;
  string user_id = 6;
  int32 notify_before_minutes = 7;
}
```

## Примечания
- Все даты/время — в формате RFC3339 (UTC).
- Для gRPC используйте proto-файл `EventService.proto`.
- Для REST используйте JSON-структуры, как в примерах. 