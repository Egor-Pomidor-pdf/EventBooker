package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/Egor-Pomidor-pdf/EventBooker/internal/models"
	"github.com/wb-go/wbf/dbpg"
)

// PostgresRepository стандартная реализация слоя repository
type PostgresRepository struct {
	db *dbpg.DB
}

func NewPostgresRepository(db *dbpg.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

func (p *PostgresRepository) CreateEvent(ctx context.Context, event *models.Event) (*models.Event, error) {
	if event == nil {
		return nil, fmt.Errorf("event не может быть nil при создании")
	}

	query := `
	insert into events (title, start_time, capacity)
	values ($1,$2,$3)
	returning id, title, start_time, capacity, created_at;
	`

	var created models.Event

	err := p.db.Master.QueryRowContext(
		ctx,
		query,
		event.Title,
		event.StartTime,
		event.Capacity,
	).Scan(
		&created.ID,
		&created.Title,
		&created.StartTime,
		&created.Capacity,
		&created.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка repository уровня при создании event: %w", err)
	}

	return &created, nil
}

func (p *PostgresRepository) GetEventByID(ctx context.Context, eventID int64) (*models.Event, error) {
	var event models.Event

	query := `
	SELECT id, title, start_time, capacity, created_at
	FROM events
	WHERE id = $1
	`

	err := p.db.Master.QueryRowContext(ctx, query, eventID).Scan(
		&event.ID,
		&event.Title,
		&event.StartTime,
		&event.Capacity,
		&event.CreatedAt,
	)
	if err == sql.ErrNoRows {
		log.Printf("запрошенный event с id %v не найден", eventID)
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка repository при получении event по id: %w", err)
	}

	return &event, nil
}

// GetAllEvents возвращает все мероприятия
func (p *PostgresRepository) GetAllEvents(ctx context.Context) ([]*models.Event, error) {
	query := `
		SELECT id, title, start_time, capacity, created_at
		FROM events
		ORDER BY start_time
	`

	rows, err := p.db.Master.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка repository при получении всех events: %w", err)
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		var e models.Event
		if err := rows.Scan(&e.ID, &e.Title, &e.StartTime, &e.Capacity, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("ошибка при сканировании row: %w", err)
		}
		events = append(events, &e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка во время перебора rows: %w", err)
	}

	return events, nil
}

// GetAllEventsWithFreePlaces возвращает будущие мероприятия и количество свободных мест
func (p *PostgresRepository) GetAllEventsWithFreePlaces(ctx context.Context) ([]models.EventWithFreePlaces, error) {
	query := `
		SELECT
			e.id, e.title, e.start_time, e.capacity, e.created_at,
			GREATEST(e.capacity - COALESCE(b.cnt, 0), 0) AS free_places
		FROM events e
		LEFT JOIN (
			SELECT event_id, COUNT(*) AS cnt
			FROM bookings
			WHERE status IN ('booked', 'confirmed')
			GROUP BY event_id
		) b ON b.event_id = e.id
		WHERE e.start_time > NOW()
		ORDER BY e.start_time
	`

	rows, err := p.db.Master.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ошибка repository при получении events с free places: %w", err)
	}
	defer rows.Close()

	var events []models.EventWithFreePlaces
	for rows.Next() {
		var e models.Event
		var free int64
		if err := rows.Scan(&e.ID, &e.Title, &e.StartTime, &e.Capacity, &e.CreatedAt, &free); err != nil {
			return nil, fmt.Errorf("ошибка при сканировании row: %w", err)
		}
		event := e
		events = append(events, models.EventWithFreePlaces{
			Event:      &event,
			FreePlaces: free,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка во время перебора rows: %w", err)
	}

	return events, nil
}

// CreateUser создаёт нового пользователя
func (p *PostgresRepository) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	if user == nil {
		return nil, fmt.Errorf("user не может быть nil")
	}

	query := `
		INSERT INTO users (name, is_admin)
		VALUES ($1, $2)
		RETURNING id, name, is_admin
	`

	var created models.User
	err := p.db.Master.QueryRowContext(ctx, query, user.Name, user.IsAdmin).Scan(
		&created.ID,
		&created.Name,
		&created.IsAdmin,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка repository при создании user: %w", err)
	}

	log.Printf("создан пользователь с ID = %v", created.ID)

	return &created, nil
}


// GetUserByID возвращает пользователя по id
func (p *PostgresRepository) GetUserByID(ctx context.Context, userID int64) (*models.User, error) {
	query := `
		SELECT id, name, is_admin
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := p.db.Master.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Name,
		&user.IsAdmin,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка repository при получении user по id: %w", err)
	}


	return &user, nil
}

// CreateBooking создаёт новую бронь, если есть свободные места
func (p *PostgresRepository) CreateBooking(ctx context.Context, booking *models.Booking) (*models.Booking, error) {
	if booking == nil {
		return nil, fmt.Errorf("booking не может быть nil")
	}

	tx, err := p.db.Master.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка начала транзакции: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	// Считаем сколько уже подтверждено/забронировано мест
	var count int64
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*) 
		FROM bookings 
		WHERE event_id = $1 AND status IN ('booked', 'confirmed')
	`, booking.EventID).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("ошибка подсчёта занятых мест: %w", err)
	}

	// Получаем вместимость события
	var capacity int64
	err = tx.QueryRowContext(ctx, `SELECT capacity FROM events WHERE id = $1 FOR UPDATE`, booking.EventID).Scan(&capacity)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения capacity события: %w", err)
	}

	if count >= capacity {
		return nil, fmt.Errorf("нет свободных мест для события %d", booking.EventID)
	}

	// Создаем бронь
	err = tx.QueryRowContext(ctx, `
		INSERT INTO bookings (event_id, user_id, status, expires_at)
		VALUES ($1, $2, 'booked', $3)
		RETURNING id, event_id, user_id, status, created_at, expires_at
	`, booking.EventID, booking.UserID, booking.ExpiresAt).Scan(
		&booking.ID,
		&booking.EventID,
		&booking.UserID,
		&booking.Status,
		&booking.CreatedAt,
		&booking.ExpiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания брони: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("ошибка коммита транзакции: %w", err)
	}

	return booking, nil
}


// GetBooking возвращает бронь по событию и пользователю
func (p *PostgresRepository) GetBooking(ctx context.Context, eventID, userID int64) (*models.Booking, error) {
	var b models.Booking
	err := p.db.Master.QueryRowContext(ctx, `
		SELECT id, event_id, user_id, status, created_at, expires_at
		FROM bookings
		WHERE event_id = $1 AND user_id = $2
	`, eventID, userID).Scan(
		&b.ID,
		&b.EventID,
		&b.UserID,
		&b.Status,
		&b.CreatedAt,
		&b.ExpiresAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка получения брони: %w", err)
	}
	return &b, nil
}

// UpdateBookingStatus обновляет статус брони
func (p *PostgresRepository) UpdateBookingStatus(ctx context.Context, bookingID int64, status string) error {
	_, err := p.db.Master.ExecContext(ctx, `
		UPDATE bookings
		SET status = $1
		WHERE id = $2
	`, status, bookingID)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса брони: %w", err)
	}
	return nil
}

// CountActiveBookings возвращает количество активных бронирований (booked + confirmed) для события
func (p *PostgresRepository) CountActiveBookings(ctx context.Context, eventID int64) (int64, error) {
	var count int64
	err := p.db.Master.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM bookings
		WHERE event_id = $1 AND status IN ('booked', 'confirmed')
	`, eventID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("ошибка подсчёта активных бронирований: %w", err)
	}
	return count, nil
}


// FindExpiredBookings возвращает все брони со статусом 'booked', у которых истёк срок
func (p *PostgresRepository) FindExpiredBookings(ctx context.Context) ([]*models.Booking, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, event_id, user_id, status, created_at, expires_at
		FROM bookings
		WHERE status = 'booked' AND expires_at < NOW()
	`)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения просроченных бронирований: %w", err)
	}
	defer rows.Close()

	var bookings []*models.Booking
	for rows.Next() {
		var b models.Booking
		if err := rows.Scan(&b.ID, &b.EventID, &b.UserID, &b.Status, &b.CreatedAt, &b.ExpiresAt); err != nil {
			return nil, fmt.Errorf("ошибка сканирования строки брони: %w", err)
		}
		bookings = append(bookings, &b)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при переборе строк бронирований: %w", err)
	}

	return bookings, nil
}


// GetBookingsByEventID возвращает все бронирования для указанного события
func (p *PostgresRepository) GetBookingsByEventID(ctx context.Context, eventID int64) ([]*models.Booking, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, event_id, user_id, status, created_at, expires_at
		FROM bookings
		WHERE event_id = $1
	`, eventID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения бронирований для события %d: %w", eventID, err)
	}
	defer rows.Close()

	var bookings []*models.Booking
	for rows.Next() {
		var b models.Booking
		if err := rows.Scan(
			&b.ID,
			&b.EventID,
			&b.UserID,
			&b.Status,
			&b.CreatedAt,
			&b.ExpiresAt,
		); err != nil {
			return nil, fmt.Errorf("ошибка сканирования брони: %w", err)
		}
		bookings = append(bookings, &b)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при переборе строк бронирований: %w", err)
	}

	return bookings, nil
}

// GetBookingByID возвращает бронь по ID
func (p *PostgresRepository) GetBookingByID(ctx context.Context, bookingID int64) (*models.Booking, error) {
	if bookingID <= 0 {
		return nil, fmt.Errorf("невалидный id брони")
	}

	var b models.Booking
	query := `
		SELECT id, event_id, user_id, status, created_at, expires_at
		FROM bookings
		WHERE id = $1
	`

	err := p.db.QueryRowContext(ctx, query, bookingID).Scan(
		&b.ID,
		&b.EventID,
		&b.UserID,
		&b.Status,
		&b.CreatedAt,
		&b.ExpiresAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("ошибка получения брони по ID: %w", err)
	}

	return &b, nil
}

// CancelExpiredBookings пометить все просроченные booked - cancelled
// и вернуть список отменённых записей.
func (p *PostgresRepository) CancelExpiredBookings(ctx context.Context) ([]*models.Booking, error) {
	tx, err := p.db.Master.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка начала транзакции: %w", err)
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
        UPDATE bookings
        SET status = 'cancelled'
        WHERE status = 'booked' AND expires_at < NOW()
        RETURNING id, event_id, user_id, status, created_at, expires_at
    `)
	if err != nil {
		return nil, fmt.Errorf("ошибка при обновлении просроченных бронирований: %w", err)
	}
	defer rows.Close()

	var cancelled []*models.Booking
	for rows.Next() {
		var b models.Booking
		if err := rows.Scan(&b.ID, &b.EventID, &b.UserID, &b.Status, &b.CreatedAt, &b.ExpiresAt); err != nil {
			return nil, fmt.Errorf("ошибка сканирования отменённой брони: %w", err)
		}
		cancelled = append(cancelled, &b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при переборе отменённых бронирований: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("ошибка коммита транзакции при отмене бронирований: %w", err)
	}

	return cancelled, nil
}

// Close закрывает соединение с базой данных
func (p *PostgresRepository) Close() error {
	if p.db != nil && p.db.Master != nil {
		return p.db.Master.Close()
	}
	return nil
}
