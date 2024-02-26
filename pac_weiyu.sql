DECLARE
    TYPE t_table_names IS TABLE OF VARCHAR2(30);
    v_table_names t_table_names := t_table_names('ACG_DLV_GEN_HIS_DBF', 'STPFC_ENTRY_TABLE', 'STPDLV_ENTRY_TABLE', 'STPDOC_ENTRY_TABLE'); -- Add table names here
BEGIN
    FOR i IN 1..v_table_names.COUNT LOOP
        EXECUTE IMMEDIATE 'ALTER TABLE ' || v_table_names(i) || ' MOVE';
        FOR j IN (SELECT index_name, table_owner FROM all_indexes WHERE table_name = UPPER(v_table_names(i))) LOOP
            EXECUTE IMMEDIATE 'ALTER INDEX ' || j.table_owner || '.' || j.index_name || ' REBUILD PARALLEL 4'; -- adjust the parallel degree on demand
            EXECUTE IMMEDIATE 'ALTER INDEX ' || j.table_owner || '.' || j.index_name || ' NOPARALLEL'; -- reset degree of parallelism to 1
        END LOOP;
    END LOOP;
END;
/