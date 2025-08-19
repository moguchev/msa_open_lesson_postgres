package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jaswdr/faker/v2"
	"github.com/moguchev/msa_open_lesson_postgres/internal/models"
	"github.com/moguchev/msa_open_lesson_postgres/internal/models/pagination"
	users_repo_bob "github.com/moguchev/msa_open_lesson_postgres/internal/repository/users/bob"
)

var (
	PostgresUser     = os.Getenv("POSTGRES_USER")
	PostgresPassword = os.Getenv("POSTGRES_PASSWORD")
	PostgresHost     = os.Getenv("POSTGRES_HOST")
	PostgresPort     = os.Getenv("POSTGRES_PORT")
	PostgresDB       = os.Getenv("POSTGRES_DB")
)

func newPostgresConnection(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		PostgresUser, PostgresPassword, PostgresHost, PostgresPort, PostgresDB,
	)
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("pgxpool connect: %w", err)
	}

	return pool, nil
}

func main() {
	ctx := context.Background()

	pool, err := newPostgresConnection(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	repo1 := users_repo_bob.NewRepository(pool)
	DemoUsersRepositoryUsage(repo1)
}

type UsersRepository interface {
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (*models.User, error)
	FindUserByEmail(ctx context.Context, email string) (*models.User, error)
	SearchUsers(ctx context.Context, filter *models.UserFilter, opts ...pagination.Option) ([]*models.User, error)
}

// DemoUsersRepositoryUsage вызывает все методы репозитория по очереди
func DemoUsersRepositoryUsage(repo UsersRepository) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	// 1. Создание пользователя
	user := &models.User{
		ID:       uuid.New(),
		Email:    RandomEmail(),
		Username: faker.New().Company().Name(),
		FullName: "Demo User",
		IsActive: true,
	}
	created, err := repo.CreateUser(ctx, user)
	if err != nil {
		log.Fatalf("CreateUser error: %v", err)
	}
	fmt.Printf("Created user: %+v\n", created)

	// 2. Получение по ID
	got, err := repo.GetUser(ctx, created.ID)
	if err != nil {
		log.Fatalf("GetUser error: %v", err)
	}
	fmt.Printf("Got user by ID: %+v\n", got)

	// 3. Поиск по email
	byEmail, err := repo.FindUserByEmail(ctx, created.Email)
	if err != nil {
		log.Fatalf("FindUserByEmail error: %v", err)
	}
	fmt.Printf("Got user by Email: %+v\n", byEmail)

	// 4. Поиск пользователей (с фильтрами + пагинацией)
	filter := &models.UserFilter{
		Email:    &created.Email,
		IsActive: &created.IsActive,
	}
	users, err := repo.SearchUsers(ctx, filter,
		pagination.WithLimit(10),
		pagination.WithOffset(0),
		pagination.WithSortFields(pagination.OrderBy("email", pagination.DESC)),
	)
	if err != nil {
		log.Fatalf("SearchUsers error: %v", err)
	}
	fmt.Printf("SearchUsers returned %d users\n", len(users))
}

var domains = []string{
	"example.com",
	"test.org",
	"demo.net",
	"mail.io",
	"sample.dev",
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandomEmail возвращает случайный email вроде "user123@test.org"
func RandomEmail() string {
	name := fmt.Sprintf("user%d", rand.Intn(1000000))
	domain := domains[rand.Intn(len(domains))]
	return fmt.Sprintf("%s@%s", name, domain)
}
