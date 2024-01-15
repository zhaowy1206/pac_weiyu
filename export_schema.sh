#!/bin/bash

# Function to extract the DDL SQL of a table and all the indexes on it
extract_ddl() {
    # The Oracle database username is the first argument to the function
    local username=$1
    # The Oracle database password is the second argument to the function
    local password=$2
    # The Oracle database connection string is the third argument to the function
    local connection_string=$3
    # The table name is the fourth argument to the function
    local table_name=$4

    # Create the SQL commands
    local sql_commands=$(cat << EOF
SET SERVEROUTPUT ON
SET HEADING OFF
SET ECHO OFF
SET PAGESIZE 0
SET TRIMSPOOL ON
SET LONG 10000
SET LONGCHUNKSIZE 10000
SET FEEDBACK OFF
SET VERIFY OFF
SET LINESIZE 32767

DECLARE
    -- Declare a variable for the DDL statement
    ddl CLOB;
BEGIN
    -- Get the DDL statement for the table
    ddl := DBMS_METADATA.GET_DDL('TABLE', '${table_name}');

    -- Output the DDL statement
    DBMS_OUTPUT.PUT_LINE(ddl);

    -- Get the DDL statements for the indexes on the table
    FOR index_record IN (SELECT index_name FROM USER_INDEXES WHERE table_name = '${table_name}' ORDER BY index_name) LOOP
        ddl := DBMS_METADATA.GET_DDL('INDEX', index_record.index_name);

        -- Output the DDL statement
        DBMS_OUTPUT.PUT_LINE(ddl);
    END LOOP;
    COMMIT;
END;
/

EXIT;
EOF
)

    # Connect to the database and run the SQL commands
    echo "$sql_commands" | sqlplus -s $username/$password@$connection_string > "${table_name}.sql" 2>&1
}

# Ask the user for the Oracle database credentials
echo "Enter your Oracle database username:"
read username
echo "Enter your Oracle database password:"
read -s password
echo "Enter your Oracle database connection string:"
read connection_string

# Get the list of tables
tables=$(echo "SET HEADING OFF FEEDBACK OFF PAGESIZE 0 VERIFY OFF LINESIZE 32767 TRIMSPOOL ON WRAP OFF;
SELECT table_name FROM USER_TABLES;" | sqlplus -s $username/$password@$connection_string)

echo "$tables" > tables.txt

# Call the function to extract the DDL SQL for each table
for table_name in $tables; do
    extract_ddl $username $password $connection_string $table_name
done

echo "Metadata exported to individual SQL files."