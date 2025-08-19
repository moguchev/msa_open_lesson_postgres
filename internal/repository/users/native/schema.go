package users

const usersTable = "users"

const (
	usersTableColumnID        = "id"
	usersTableColumnEmail     = "email"
	usersTableColumnUsername  = "username"
	usersTableColumnFullname  = "full_name"
	usersTableColumnCreatedAt = "created_at"
	usersTableColumnLastLogin = "last_login"
	usersTableColumnIsActive  = "is_active"
)

var usersTableColumns = []string{
	usersTableColumnID,
	usersTableColumnEmail,
	usersTableColumnUsername,
	usersTableColumnFullname,
	usersTableColumnCreatedAt,
	usersTableColumnLastLogin,
	usersTableColumnIsActive,
}
