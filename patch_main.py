with open("cmd/server/main.go", "r") as f:
    content = f.read()

import_cron = '"github.com/cecep-azhar/jurnalumi/internal/scheduler"'
if import_cron not in content:
    content = content.replace('"github.com/cecep-azhar/jurnalumi/internal/handlers"', '"github.com/cecep-azhar/jurnalumi/internal/handlers"\n\t"github.com/cecep-azhar/jurnalumi/internal/scheduler"')

init_cron = """	// Connect Database & Migrate
	db.InitDB(dsn)
	
	// Start Scheduler
	scheduler.InitScheduler()"""
content = content.replace("	// Connect Database & Migrate\n\tdb.InitDB(dsn)", init_cron)

with open("cmd/server/main.go", "w") as f:
    f.write(content)
