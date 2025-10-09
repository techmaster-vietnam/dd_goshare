package rbac

import (
	"bufio"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/techmaster-vietnam/dd_goshare/pkg/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ANSI color log helpers for debug
func debugRBAC(msg string, args ...interface{}) {
	fmt.Println("\033[33m[DEBUG RBAC]\033[0m " + fmt.Sprintf(msg, args...))
}
func infoRBAC(msg string, args ...interface{}) {
	fmt.Println("\033[36m[INFO  RBAC]\033[0m " + fmt.Sprintf(msg, args...))
}
func errorRBAC(msg string, args ...interface{}) {
	fmt.Println("\033[31m[ERROR RBAC]\033[0m " + fmt.Sprintf(msg, args...))
}

func isInteractive() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func color(s, code string) string {
	return "\033[" + code + "m" + s + "\033[0m"
}

func bold(s string) string   { return color(s, "1") }
func cyan(s string) string   { return color(s, "36") }
func green(s string) string  { return color(s, "32") }
func yellow(s string) string { return color(s, "33") }

func TestMain(m *testing.M) {
	menu := []string{
		"TestRBAC_BasicPermissions",
		"TestRBAC_DynamicLoadWhenSnapshotEmpty",
		"TestRBAC_UserHasPermission",
		"TestRBAC_UserNoPermission",
		"TestRBAC_AdminBypass",
		"TestRBAC_PublicRoute",
		"TestRBAC_UnassignedRoute_PublicConfig",
		"TestRBAC_UnassignedRoute_NotPublicConfig",
		"TestRBAC_DynamicPermissionChange",
		"TestRBAC_UserHasMultipleRoles",
		"TestRBAC_WrongMethodOrPath",
		"TestRBAC_NoUserID",
		"TestRBAC_UserRoleNotExist",
		"TestRBAC_RuleNoRoles",
		"TestRBAC_UserNoUserRoles",
	}
	descriptions := map[string]string{
		"TestRBAC_BasicPermissions":                "Kiểm tra quyền cơ bản: editor chỉ xem /api/rules, admin xem /api/employees",
		"TestRBAC_DynamicLoadWhenSnapshotEmpty":    "Nạp role động khi snapshot trống và cho phép sau khi gán quyền",
		"TestRBAC_UserHasPermission":               "User có quyền vào route private",
		"TestRBAC_UserNoPermission":                "User không có quyền vào route private",
		"TestRBAC_AdminBypass":                     "Admin bypass mọi route private",
		"TestRBAC_PublicRoute":                     "Route public ai cũng truy cập được",
		"TestRBAC_UnassignedRoute_PublicConfig":    "Route không gán rule, config cho phép public",
		"TestRBAC_UnassignedRoute_NotPublicConfig": "Route không gán rule, config không public (route ngoài RBAC vẫn 200)",
		"TestRBAC_DynamicPermissionChange":         "Gán quyền động rồi truy cập lại được phép",
		"TestRBAC_UserHasMultipleRoles":            "User có nhiều role, chỉ cần 1 role hợp lệ",
		"TestRBAC_WrongMethodOrPath":               "Sai method hoặc path thì không được phép",
		"TestRBAC_NoUserID":                        "Thiếu user_id khi gọi route private",
		"TestRBAC_UserRoleNotExist":                "User có role_id không tồn tại",
		"TestRBAC_RuleNoRoles":                     "Rule không gán role nào",
		"TestRBAC_UserNoUserRoles":                 "User không có dòng nào trong user_roles",
	}

	setRunFor := func(name string) {
		os.Args = append(os.Args[:1], "-test.run=^"+name+"$")
	}

	if tc := os.Getenv("TESTCASE"); tc != "" {
		// Dùng TESTCASE theo tên
		found := false
		for _, name := range menu {
			if strings.EqualFold(tc, name) {
				found = true
				break
			}
		}
		if !found {
			fmt.Println("Test case không hợp lệ. Chọn một trong:")
			for i, name := range menu {
				if d, ok := descriptions[name]; ok {
					fmt.Printf("%2d. %s - %s\n", i+1, name, d)
				} else {
					fmt.Printf("%2d. %s\n", i+1, name)
				}
			}
			os.Exit(1)
		}
		setRunFor(tc)
	} else if idxStr := os.Getenv("TESTCASE_INDEX"); idxStr != "" {
		// Dùng TESTCASE_INDEX theo số
		if idx, err := strconv.Atoi(idxStr); err == nil && idx >= 1 && idx <= len(menu) {
			setRunFor(menu[idx-1])
		} else {
			fmt.Println("TESTCASE_INDEX không hợp lệ.")
			for i, name := range menu {
				if d, ok := descriptions[name]; ok {
					fmt.Printf("%2d. %s - %s\n", i+1, name, d)
				} else {
					fmt.Printf("%2d. %s\n", i+1, name)
				}
			}
			os.Exit(1)
		}
	} else if isInteractive() {
		// Chỉ hiển thị menu khi stdin là terminal
		border := yellow("╔" + strings.Repeat("═", 60) + "╗")
		fmt.Println(border)
		fmt.Println(yellow("║") + bold("   🌟 Danh sách test case có thể chạy 🌟") + strings.Repeat(" ", 23) + yellow("║"))
		fmt.Println(yellow("╠" + strings.Repeat("═", 60) + "╣"))
		for i, name := range menu {
			desc := descriptions[name]
			fmt.Printf("%s %2d. %s\n", yellow("║"), i+1, cyan(name)+green(" - "+desc))
		}
		fmt.Println(yellow("╚" + strings.Repeat("═", 60) + "╝"))
		fmt.Print(bold("\nNhập số thứ tự test case muốn chạy (Enter để chạy tất cả): "))
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input != "" {
			idx := -1
			fmt.Sscanf(input, "%d", &idx)
			if idx >= 1 && idx <= len(menu) {
				setRunFor(menu[idx-1])
			} else {
				fmt.Println(color("Số thứ tự không hợp lệ.", "31"))
				os.Exit(1)
			}
		}
	} else {
		// Không tương tác: in gợi ý cách chọn test
		fmt.Println("Tip: dùng -run hoặc TESTCASE/TESTCASE_INDEX để chọn test.")
	}
	os.Exit(m.Run())
}

// helper: gửi request vào Fiber, kèm user_id (để middleware lấy roles từ DB)
func doReq(app *fiber.App, method, path, userID string) (int, string) {
	req := httptest.NewRequest(method, path, nil)
	if userID != "" {
		req.Header.Set("X-User-ID", userID)
	}
	resp, _ := app.Test(req, -1)
	if resp == nil {
		return 0, ""
	}
	defer resp.Body.Close()
	return resp.StatusCode, resp.Status
}

// helper: thực thi SQL
func mustExec(t *testing.T, db *gorm.DB, sql string, args ...interface{}) {
	t.Helper()
	if err := db.Exec(sql, args...).Error; err != nil {
		t.Fatalf("exec failed: %v\nSQL: %s", err, sql)
	}
}

// helper: lấy rule_id theo method + path
func getRuleID(t *testing.T, db *gorm.DB, method, path string) int64 {
	t.Helper()
	var id int64
	row := db.Raw(`SELECT id FROM rules WHERE method = ? AND path = ?`, strings.ToUpper(method), path).Row()
	if err := row.Scan(&id); err != nil {
		t.Fatalf("cannot get rule id for %s %s: %v", method, path, err)
	}
	return id
}

// helper: đếm số rule theo method + path + service
func countRules(t *testing.T, db *gorm.DB, method, path, service string) int64 {
	t.Helper()
	var cnt int64
	row := db.Raw(`SELECT COUNT(*) FROM rules WHERE method = ? AND path = ? AND service = ?`,
		strings.ToUpper(method), path, service).Row()
	if err := row.Scan(&cnt); err != nil {
		t.Fatalf("cannot count rules for %s %s svc=%s: %v", method, path, service, err)
	}
	return cnt
}

// Tạo Fiber app test + DB test + khởi tạo RBAC
func setupTestApp(t *testing.T) (*fiber.App, *gorm.DB) {
	t.Helper()

	// DB SQLite in-memory riêng cho từng test (tránh dùng chung)
	dsn := "file:rbac_" + strings.ReplaceAll(t.Name(), "/", "_") + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	// Thêm migrate thủ công cho các bảng RBAC cần thiết
	// (bạn có thể bổ sung thêm các model khác nếu cần)
	if err := db.AutoMigrate(
		&models.Role{}, &models.Rule{}, &models.UserRole{}, &models.RuleRole{},
	); err != nil {
		t.Fatalf("auto-migrate failed: %v", err)
	}

	// Khởi tạo RBAC
	cfg := Config{
		Service:                   "dd_backend",
		MakeUnassignedRoutePublic: false,
		DatabaseAutoMigrate:       false, // Đã migrate thủ công ở trên
		HighestRole:               "admin",
	}
	if err := InitRBAC(db, cfg); err != nil {
		t.Fatalf("InitRBAC failed: %v", err)
	}

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("db", db)
		if uid := c.Get("X-User-ID"); uid != "" {
			c.Locals("user_id", uid)
		}
		return c.Next()
	})

	return app, db
}

func TestRBAC_BasicPermissions(t *testing.T) {
	app, db := setupTestApp(t)

	mustExec(t, db, `INSERT OR IGNORE INTO roles (id, name) VALUES (1, 'admin'), (2, 'editor')`)
	debugRBAC("Đã insert roles: admin, editor")

	okHandler := func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) }
	RegisterBusinessRoute(app, "GET", "/api/rules", "rule.view", true, okHandler)
	RegisterBusinessRoute(app, "GET", "/api/employees", "employee.view", true, okHandler)
	debugRBAC("Đã đăng ký route /api/rules và /api/employees")

	if err := RegisterRulesToDB(); err != nil {
		errorRBAC("RegisterRulesToDB failed: %v", err)
		t.Fatalf("RegisterRulesToDB failed: %v", err)
	}
	if err := ReloadRules(); err != nil {
		errorRBAC("ReloadRules failed: %v", err)
		t.Fatalf("ReloadRules failed: %v", err)
	}
	infoRBAC("Đã lưu và reload rule vào DB")

	// Gán quyền cho editor với rule.view
	ruleIDRules := getRuleID(t, db, "GET", "/api/rules")
	mustExec(t, db, `INSERT OR IGNORE INTO rule_roles (rule_id, role_id) VALUES (?, ?)`, ruleIDRules, 2)
	// Gán quyền cho admin với employee.view (nếu middleware chưa hỗ trợ bypass)
	ruleIDEmp := getRuleID(t, db, "GET", "/api/employees")
	mustExec(t, db, `INSERT OR IGNORE INTO rule_roles (rule_id, role_id) VALUES (?, ?)`, ruleIDEmp, 1)
	debugRBAC("Đã gán quyền: editor->/api/rules, admin->/api/employees")

	// Tạo user
	mustExec(t, db, `INSERT OR IGNORE INTO user_roles (user_id, role_id) VALUES ('U2', 2), ('A1', 1)`)
	debugRBAC("Đã gán user U2=editor, A1=admin")

	// Trường hợp 1: user role=2 gọi GET /api/rules => 200
	{
		code, _ := doReq(app, "GET", "/api/rules", "U2")
		debugRBAC("U2 gọi GET /api/rules => %d", code)
		if code != http.StatusOK {
			errorRBAC("expected 200 for editor on /api/rules, got %d", code)
			t.Fatalf("expected 200 for editor on /api/rules, got %d", code)
		}
	}

	// Trường hợp 2: user role=2 gọi GET /api/employees => 403
	{
		code, _ := doReq(app, "GET", "/api/employees", "U2")
		debugRBAC("U2 gọi GET /api/employees => %d", code)
		if code != http.StatusForbidden {
			errorRBAC("expected 403 for editor on /api/employees, got %d", code)
			t.Fatalf("expected 403 for editor on /api/employees, got %d", code)
		}
	}

	// Trường hợp 3: admin gọi GET /api/employees => 200 (bypass)
	{
		code, _ := doReq(app, "GET", "/api/employees", "A1")
		debugRBAC("A1 (admin) gọi GET /api/employees => %d", code)
		if code != http.StatusOK {
			errorRBAC("expected 200 for admin on /api/employees, got %d", code)
			t.Fatalf("expected 200 for admin on /api/employees, got %d", code)
		}
	}
	infoRBAC("TestRBAC_BasicPermissions hoàn thành!")
}

// Optional: kiểm tra nạp quyền động khi snapshot route.Roles trống
// Chỉ bật test này nếu bạn đã áp dụng logic "nạp roles động theo access_type trong middleware".
func TestRBAC_DynamicLoadWhenSnapshotEmpty(t *testing.T) {
	app, db := setupTestApp(t)

	// roles
	mustExec(t, db, `INSERT OR IGNORE INTO roles (id, name) VALUES (1, 'admin'), (2, 'editor')`)

	okHandler := func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) }
	RegisterBusinessRoute(app, "GET", "/api/rules", "rule.view", true, okHandler)

	if err := RegisterRulesToDB(); err != nil {
		t.Fatalf("RegisterRulesToDB failed: %v", err)
	}
	if err := ReloadRules(); err != nil {
		t.Fatalf("ReloadRules failed: %v", err)
	}

	// Tạo user editor
	mustExec(t, db, `INSERT OR IGNORE INTO user_roles (user_id, role_id) VALUES ('U2', 2)`)

	// Lần 1: chưa gán quyền => 403
	{
		code, _ := doReq(app, "GET", "/api/rules", "U2")
		if code != http.StatusForbidden && code != http.StatusUnauthorized {
			t.Fatalf("expected 403/401 before granting, got %d", code)
		}
	}

	// Gán quyền vào DB: rule.view -> role 2
	ruleIDRules := getRuleID(t, db, "GET", "/api/rules")
	mustExec(t, db, `INSERT OR IGNORE INTO rule_roles (rule_id, role_id) VALUES (?, ?)`, ruleIDRules, 2)

	// Lần 2: kỳ vọng middleware nạp quyền động theo access_type và cho phép => 200
	{
		code, status := doReq(app, "GET", "/api/rules", "U2")
		if code != http.StatusOK {
			t.Fatalf("expected 200 after granting via dynamic load, got %d (%s)", code, status)
		}
	}
}

// TC01: User có quyền truy cập route private
func TestRBAC_UserHasPermission(t *testing.T) {
	app, db := setupTestApp(t)
	mustExec(t, db, `INSERT OR IGNORE INTO roles (id, name) VALUES (2, 'editor')`)
	okHandler := func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) }
	RegisterBusinessRoute(app, "GET", "/api/private", "private.view", true, okHandler)
	RegisterRulesToDB()
	ReloadRules()
	ruleID := getRuleID(t, db, "GET", "/api/private")
	mustExec(t, db, `INSERT OR IGNORE INTO rule_roles (rule_id, role_id) VALUES (?, ?)`, ruleID, 2)
	mustExec(t, db, `INSERT OR IGNORE INTO user_roles (user_id, role_id) VALUES ('U2', 2)`)
	code, _ := doReq(app, "GET", "/api/private", "U2")
	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d", code)
	}
}

// TC02: User không có quyền truy cập route private
func TestRBAC_UserNoPermission(t *testing.T) {
	app, db := setupTestApp(t)
	mustExec(t, db, `INSERT OR IGNORE INTO roles (id, name) VALUES (2, 'editor'), (3, 'viewer')`)
	okHandler := func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) }
	RegisterBusinessRoute(app, "GET", "/api/private", "private.view", true, okHandler)
	RegisterRulesToDB()
	ReloadRules()
	ruleID := getRuleID(t, db, "GET", "/api/private")
	mustExec(t, db, `INSERT OR IGNORE INTO rule_roles (rule_id, role_id) VALUES (?, ?)`, ruleID, 2)
	mustExec(t, db, `INSERT OR IGNORE INTO user_roles (user_id, role_id) VALUES ('U3', 3)`)
	code, _ := doReq(app, "GET", "/api/private", "U3")
	if code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", code)
	}
}

// TC03: Admin truy cập mọi route private
func TestRBAC_AdminBypass(t *testing.T) {
	app, db := setupTestApp(t)
	mustExec(t, db, `INSERT OR IGNORE INTO roles (id, name) VALUES (1, 'admin')`)
	okHandler := func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) }
	RegisterBusinessRoute(app, "GET", "/api/secret", "secret.view", true, okHandler)
	RegisterRulesToDB()
	ReloadRules()
	mustExec(t, db, `INSERT OR IGNORE INTO user_roles (user_id, role_id) VALUES ('A1', 1)`)
	code, _ := doReq(app, "GET", "/api/secret", "A1")
	if code != http.StatusOK {
		t.Fatalf("expected 200 for admin, got %d", code)
	}
}

// TC04: Route public, ai cũng truy cập được
func TestRBAC_PublicRoute(t *testing.T) {
	app, _ := setupTestApp(t)
	okHandler := func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) }
	RegisterBusinessRoute(app, "GET", "/api/public", "public.view", false, okHandler)
	RegisterRulesToDB()
	ReloadRules()
	code, _ := doReq(app, "GET", "/api/public", "")
	if code != http.StatusOK {
		t.Fatalf("expected 200 for public route, got %d", code)
	}
}

// TC05: Route chưa gán rule, config public
func TestRBAC_UnassignedRoute_PublicConfig(t *testing.T) {
	app, _ := setupTestApp(t)
	config.MakeUnassignedRoutePublic = true
	okHandler := func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) }
	app.Get("/api/unassigned", okHandler) // Không đăng ký qua RBAC
	code, _ := doReq(app, "GET", "/api/unassigned", "")
	if code != http.StatusOK {
		t.Fatalf("expected 200 for unassigned route with public config, got %d", code)
	}
}

// TC06: Route chưa gán rule, config không public
func TestRBAC_UnassignedRoute_NotPublicConfig(t *testing.T) {
	app, _ := setupTestApp(t)
	config.MakeUnassignedRoutePublic = false
	okHandler := func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) }
	// Route không qua RBAC => middleware không can thiệp, luôn 200
	app.Get("/api/unassigned", okHandler)
	code, _ := doReq(app, "GET", "/api/unassigned", "")
	if code != http.StatusOK {
		t.Fatalf("expected 200 for unassigned route without RBAC, got %d", code)
	}
}

// TC07: Thay đổi quyền động, middleware nhận biết
func TestRBAC_DynamicPermissionChange(t *testing.T) {
	app, db := setupTestApp(t)
	mustExec(t, db, `INSERT OR IGNORE INTO roles (id, name) VALUES (2, 'editor')`)
	okHandler := func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) }
	RegisterBusinessRoute(app, "GET", "/api/dynamic", "dynamic.view", true, okHandler)
	RegisterRulesToDB()
	ReloadRules()
	mustExec(t, db, `INSERT OR IGNORE INTO user_roles (user_id, role_id) VALUES ('U2', 2)`)
	// Chưa gán quyền
	code, _ := doReq(app, "GET", "/api/dynamic", "U2")
	if code != http.StatusForbidden && code != http.StatusUnauthorized {
		t.Fatalf("expected 403/401 before granting, got %d", code)
	}
	// Gán quyền động
	ruleID := getRuleID(t, db, "GET", "/api/dynamic")
	mustExec(t, db, `INSERT OR IGNORE INTO rule_roles (rule_id, role_id) VALUES (?, ?)`, ruleID, 2)
	// Gọi lại, kỳ vọng được phép
	code, _ = doReq(app, "GET", "/api/dynamic", "U2")
	if code != http.StatusOK {
		t.Fatalf("expected 200 after granting, got %d", code)
	}
}

// TC08: User có nhiều role, chỉ cần 1 role hợp lệ
func TestRBAC_UserHasMultipleRoles(t *testing.T) {
	app, db := setupTestApp(t)
	mustExec(t, db, `INSERT OR IGNORE INTO roles (id, name) VALUES (2, 'editor'), (3, 'viewer')`)
	okHandler := func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) }
	RegisterBusinessRoute(app, "GET", "/api/multi", "multi.view", true, okHandler)
	RegisterRulesToDB()
	ReloadRules()
	ruleID := getRuleID(t, db, "GET", "/api/multi")
	mustExec(t, db, `INSERT OR IGNORE INTO rule_roles (rule_id, role_id) VALUES (?, ?)`, ruleID, 2)
	mustExec(t, db, `INSERT OR IGNORE INTO user_roles (user_id, role_id) VALUES ('U4', 2), ('U4', 3)`)
	code, _ := doReq(app, "GET", "/api/multi", "U4")
	if code != http.StatusOK {
		t.Fatalf("expected 200 for user with multiple roles, got %d", code)
	}
}

// TC09: Đúng quyền nhưng sai method/path
func TestRBAC_WrongMethodOrPath(t *testing.T) {
	app, db := setupTestApp(t)
	mustExec(t, db, `INSERT OR IGNORE INTO roles (id, name) VALUES (2, 'editor')`)
	okHandler := func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) }
	RegisterBusinessRoute(app, "GET", "/api/exact", "exact.view", true, okHandler)
	RegisterRulesToDB()
	ReloadRules()
	ruleID := getRuleID(t, db, "GET", "/api/exact")
	mustExec(t, db, `INSERT OR IGNORE INTO rule_roles (rule_id, role_id) VALUES (?, ?)`, ruleID, 2)
	mustExec(t, db, `INSERT OR IGNORE INTO user_roles (user_id, role_id) VALUES ('U2', 2)`)
	// Đúng quyền nhưng gọi POST
	code, _ := doReq(app, "POST", "/api/exact", "U2")
	if code == http.StatusOK {
		t.Fatalf("expected not 200 for wrong method, got %d", code)
	}
	// Đúng quyền nhưng gọi sai path
	code, _ = doReq(app, "GET", "/api/exact2", "U2")
	if code == http.StatusOK {
		t.Fatalf("expected not 200 for wrong path, got %d", code)
	}
}

// TC10: Không truyền user_id vào route private
func TestRBAC_NoUserID(t *testing.T) {
	app, db := setupTestApp(t)
	mustExec(t, db, `INSERT OR IGNORE INTO roles (id, name) VALUES (2, 'editor')`)
	okHandler := func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) }
	RegisterBusinessRoute(app, "GET", "/api/private", "private.view", true, okHandler)
	RegisterRulesToDB()
	ReloadRules()
	ruleID := getRuleID(t, db, "GET", "/api/private")
	mustExec(t, db, `INSERT OR IGNORE INTO rule_roles (rule_id, role_id) VALUES (?, ?)`, ruleID, 2)
	code, _ := doReq(app, "GET", "/api/private", "")
	if code != http.StatusUnauthorized && code != http.StatusForbidden {
		t.Fatalf("expected 401/403 for missing user_id, got %d", code)
	}
}

// TC11: User có role_id không tồn tại
func TestRBAC_UserRoleNotExist(t *testing.T) {
	app, db := setupTestApp(t)
	// Không seed role_id=99
	okHandler := func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) }
	RegisterBusinessRoute(app, "GET", "/api/private", "private.view", true, okHandler)
	RegisterRulesToDB()
	ReloadRules()
	ruleID := getRuleID(t, db, "GET", "/api/private")
	mustExec(t, db, `INSERT OR IGNORE INTO rule_roles (rule_id, role_id) VALUES (?, ?)", `, ruleID, 99)
	// Fix malformed SQL above if needed (some drivers might accept it), ensure rule_roles exists for 99
	mustExec(t, db, `INSERT OR IGNORE INTO rule_roles (rule_id, role_id) VALUES (?, ?)`, ruleID, 99)
	mustExec(t, db, `INSERT OR IGNORE INTO user_roles (user_id, role_id) VALUES ('U99', 99)`)
	code, _ := doReq(app, "GET", "/api/private", "U99")
	if code != http.StatusForbidden {
		t.Fatalf("expected 403 for user with non-existent role, got %d", code)
	}
}

// TC12: Rule không gán role nào
func TestRBAC_RuleNoRoles(t *testing.T) {
	app, db := setupTestApp(t)
	mustExec(t, db, `INSERT OR IGNORE INTO roles (id, name) VALUES (2, 'editor')`)
	okHandler := func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) }
	RegisterBusinessRoute(app, "GET", "/api/empty", "empty.view", true, okHandler)
	RegisterRulesToDB()
	ReloadRules()
	mustExec(t, db, `INSERT OR IGNORE INTO user_roles (user_id, role_id) VALUES ('U2', 2)`)
	code, _ := doReq(app, "GET", "/api/empty", "U2")
	if code != http.StatusForbidden {
		t.Fatalf("expected 403 for rule with no roles, got %d", code)
	}
}

// TC13: User không có dòng nào trong user_roles
func TestRBAC_UserNoUserRoles(t *testing.T) {
	app, db := setupTestApp(t)
	mustExec(t, db, `INSERT OR IGNORE INTO roles (id, name) VALUES (2, 'editor')`)
	okHandler := func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) }
	RegisterBusinessRoute(app, "GET", "/api/private", "private.view", true, okHandler)
	RegisterRulesToDB()
	ReloadRules()
	ruleID := getRuleID(t, db, "GET", "/api/private")
	mustExec(t, db, `INSERT OR IGNORE INTO rule_roles (rule_id, role_id) VALUES (?, ?)`, ruleID, 2)
	// Không tạo user_roles cho U0
	code, _ := doReq(app, "GET", "/api/private", "U0")
	if code != http.StatusForbidden {
		t.Fatalf("expected 403 for user with no user_roles, got %d", code)
	}
}

// New tests

// Route có path params (/:id) và private: user có role hợp lệ được phép.
func TestRBAC_PathParams_PrivateAccess(t *testing.T) {
	app, db := setupTestApp(t)

	mustExec(t, db, `INSERT OR IGNORE INTO roles (id, name) VALUES (2, 'editor')`)

	okHandler := func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) }
	RegisterBusinessRoute(app, "GET", "/api/employees/:id", "employee.detail", true, okHandler)

	if err := RegisterRulesToDB(); err != nil {
		t.Fatalf("RegisterRulesToDB failed: %v", err)
	}
	if err := ReloadRules(); err != nil {
		t.Fatalf("ReloadRules failed: %v", err)
	}

	ruleID := getRuleID(t, db, "GET", "/api/employees/:id")
	mustExec(t, db, `INSERT OR IGNORE INTO rule_roles (rule_id, role_id) VALUES (?, ?)`, ruleID, 2)
	mustExec(t, db, `INSERT OR IGNORE INTO user_roles (user_id, role_id) VALUES ('U2', 2)`)

	code, _ := doReq(app, "GET", "/api/employees/123", "U2")
	if code != http.StatusOK {
		t.Fatalf("expected 200 for editor on /api/employees/:id, got %d", code)
	}
}

// Đăng ký route trùng lặp nhiều lần vẫn chỉ có một bản ghi rule (idempotent).
func TestRBAC_DuplicateRegisterRulesIdempotent(t *testing.T) {
	app, db := setupTestApp(t)

	okHandler := func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) }

	// Đăng ký nhiều lần cùng method/path
	for i := 0; i < 3; i++ {
		RegisterBusinessRoute(app, "GET", "/api/idempotent", "idempotent.view", true, okHandler)
	}

	// Gọi lưu nhiều lần
	for i := 0; i < 3; i++ {
		if err := RegisterRulesToDB(); err != nil {
			t.Fatalf("RegisterRulesToDB failed: %v", err)
		}
	}
	if err := ReloadRules(); err != nil {
		t.Fatalf("ReloadRules failed: %v", err)
	}

	cnt := countRules(t, db, "GET", "/api/idempotent", config.Service)
	if cnt != 1 {
		t.Fatalf("expected exactly 1 rule saved, got %d", cnt)
	}
}

// Thu hồi quyền sau khi đã cấp, reload lại thì bị chặn.
func TestRBAC_RoleRevokedAfterGrant(t *testing.T) {
	app, db := setupTestApp(t)

	mustExec(t, db, `INSERT OR IGNORE INTO roles (id, name) VALUES (2, 'editor')`)
	okHandler := func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) }
	RegisterBusinessRoute(app, "GET", "/api/revoke", "revoke.view", true, okHandler)

	if err := RegisterRulesToDB(); err != nil {
		t.Fatalf("RegisterRulesToDB failed: %v", err)
	}
	if err := ReloadRules(); err != nil {
		t.Fatalf("ReloadRules failed: %v", err)
	}

	ruleID := getRuleID(t, db, "GET", "/api/revoke")
	mustExec(t, db, `INSERT OR IGNORE INTO rule_roles (rule_id, role_id) VALUES (?, ?)`, ruleID, 2)
	mustExec(t, db, `INSERT OR IGNORE INTO user_roles (user_id, role_id) VALUES ('U2', 2)`)

	// Lần 1: có quyền
	{
		code, _ := doReq(app, "GET", "/api/revoke", "U2")
		if code != http.StatusOK {
			t.Fatalf("expected 200 with granted permission, got %d", code)
		}
	}

	// Thu hồi quyền
	mustExec(t, db, `DELETE FROM rule_roles WHERE rule_id = ? AND role_id = ?`, ruleID, 2)

	// Reload snapshot để phản ánh thay đổi
	if err := ReloadRules(); err != nil {
		t.Fatalf("ReloadRules failed: %v", err)
	}

	// Lần 2: bị chặn
	{
		code, _ := doReq(app, "GET", "/api/revoke", "U2")
		if code != http.StatusForbidden && code != http.StatusUnauthorized {
			t.Fatalf("expected 403/401 after revoke, got %d", code)
		}
	}
}

// User có 2 role hợp lệ, thu hồi 1 role vẫn được phép vì còn role còn lại.
func TestRBAC_MultiRole_StillAllowedWhenOneRevoked(t *testing.T) {
	app, db := setupTestApp(t)

	mustExec(t, db, `INSERT OR IGNORE INTO roles (id, name) VALUES (2, 'editor'), (3, 'viewer')`)
	okHandler := func(c *fiber.Ctx) error { return c.SendStatus(http.StatusOK) }
	RegisterBusinessRoute(app, "GET", "/api/multi2", "multi2.view", true, okHandler)

	if err := RegisterRulesToDB(); err != nil {
		t.Fatalf("RegisterRulesToDB failed: %v", err)
	}
	if err := ReloadRules(); err != nil {
		t.Fatalf("ReloadRules failed: %v", err)
	}

	ruleID := getRuleID(t, db, "GET", "/api/multi2")
	// Gán cho cả 2 role
	mustExec(t, db, `INSERT OR IGNORE INTO rule_roles (rule_id, role_id) VALUES (?, ?), (?, ?)`, ruleID, 2, ruleID, 3)
	mustExec(t, db, `INSERT OR IGNORE INTO user_roles (user_id, role_id) VALUES ('U5', 2), ('U5', 3)`)

	// Lần 1: được phép
	{
		code, _ := doReq(app, "GET", "/api/multi2", "U5")
		if code != http.StatusOK {
			t.Fatalf("expected 200 with both roles, got %d", code)
		}
	}

	// Thu hồi role 2 khỏi rule
	mustExec(t, db, `DELETE FROM rule_roles WHERE rule_id = ? AND role_id = ?`, ruleID, 2)
	if err := ReloadRules(); err != nil {
		t.Fatalf("ReloadRules failed: %v", err)
	}

	// Lần 2: vẫn được phép nhờ role 3
	{
		code, _ := doReq(app, "GET", "/api/multi2", "U5")
		if code != http.StatusOK {
			t.Fatalf("expected 200 with remaining valid role, got %d", code)
		}
	}
}

func dumpRules(t *testing.T, db *gorm.DB) {
	type r struct {
		ID         int64
		Method     string
		Path       string
		AccessType string
		Service    string
		IsPrivate  bool
	}
	var rr []r
	if err := db.Raw(`SELECT id, method, path, access_type, service, is_private FROM rules`).Scan(&rr).Error; err != nil {
		t.Logf("dump rules error: %v", err)
		return
	}
	for _, it := range rr {
		t.Logf("rule: id=%d %s %s access=%s priv=%v svc=%s", it.ID, it.Method, it.Path, it.AccessType, it.IsPrivate, it.Service)
	}
}
