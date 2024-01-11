#!/bin/bash

# Ask the user for the schema name
echo "Enter the schema name:"
read schema_name

# Create the SQL commands
sql_commands=$(cat << EOF
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
    -- Declare a cursor for the objects
    CURSOR object_cursor IS
        SELECT object_type, object_name
        FROM USER_OBJECTS
        WHERE object_type IN ('TABLE', 'INDEX');

    -- Declare a record for the object metadata
    object_record object_cursor%ROWTYPE;

    -- Declare a variable for the DDL statement
    ddl CLOB;
BEGIN
    EXECUTE IMMEDIATE 'CREATE TABLE temp_metadata (
        object_type VARCHAR2(19),
        object_name VARCHAR2(128),
        ddl CLOB
    )';

    -- Open the cursor
    OPEN object_cursor;

    -- Fetch the objects into the record
    LOOP
        FETCH object_cursor INTO object_record;
        EXIT WHEN object_cursor%NOTFOUND;

        -- Get the DDL statement
        ddl := DBMS_METADATA.GET_DDL(object_record.object_type, object_record.object_name);

        -- Insert the object metadata into the temporary table
        EXECUTE IMMEDIATE 'INSERT INTO temp_metadata VALUES (:1, :2, :3)' USING
            object_record.object_type,
            object_record.object_name,
            ddl;
        COMMIT;
    END LOOP;

    -- Close the cursor
    CLOSE object_cursor;
    COMMIT;
END;
/

SPOOL ${schema_name}_metadata.txt
SELECT ddl FROM temp_metadata order by object_type, object_name;
SPOOL OFF

DROP TABLE temp_metadata;

EXIT;
EOF
)

# Connect to the database and run the SQL commands
echo "Enter your Oracle database username:"
read username
echo "Enter your Oracle database password:"
read -s password
echo "Enter your Oracle database connection string:"
read connection_string

echo "$sql_commands" | sqlplus -s $username/$password@$connection_string > /dev/null

echo "Metadata exported to ${schema_name}_metadata.txt."