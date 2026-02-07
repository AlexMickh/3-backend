import sys
import sqlite3
from datetime import datetime

def main():
    if len(sys.argv) != 2:
        print("need db path")
        sys.exit(1)

    with sqlite3.connect(sys.argv[1]) as connection:
        cursor = connection.cursor()

        cursor.execute(
            "UPDATE products SET discount = 0, discount_expires_at = NULL WHERE discount_expires_at > ?", 
            (datetime.now().isoformat(),)
        )

        connection.commit()

if __name__ == "__main__":
    main()
