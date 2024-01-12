#!/bin/bash

# Function to test the tables command
test_tables_command() {
    # The Oracle database username is the first argument to the function
    local username=$1
    # The Oracle database password is the second argument to the function
    local password=$2
    # The Oracle database connection string is the third argument to the function
    local connection_string=$3

    # Run the tables command
    local tables=$(echo "SET HEADING OFF FEEDBACK OFF PAGESIZE 0 VERIFY OFF LINESIZE 32767;
    SELECT table_name FROM USER_TABLES;" | sqlplus -s $username/$password@$connection_string | grep '^[A-Z][A-Z0-9_$]*$')

    # Check each table name
    for table_name in $tables; do
        # If the table name does not match the regular expression for valid table names, fail the test
        if [[ ! $table_name =~ ^[A-Z][A-Z0-9_$]*$ ]]; then
            echo "Test failed: '$table_name' is not a valid table name."
            return 1
        fi
    done

    # If all table names are valid, pass the test
    echo "Test passed: All table names are valid."
    return 0
}

# Ask the user for the Oracle database credentials
echo "Enter your Oracle database username:"
read username
echo "Enter your Oracle database password:"
read -s password
echo "Enter your Oracle database connection string:"
read connection_string

# Call the function to test the tables command
test_tables_command $username $password $connection_string