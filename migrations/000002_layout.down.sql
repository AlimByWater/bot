DROP INDEX IF EXISTS elysium.idx_layout_changes_timestamp;
DROP INDEX IF EXISTS elysium.idx_layout_changes_user_id;
DROP INDEX IF EXISTS elysium.idx_layout_changes_layout_id;

-- Удаление таблицы
DROP TABLE IF EXISTS elysium.layout_changes;

-- Удаление индексов
DROP INDEX IF EXISTS  elysium.idx_layout_elements_properties;
DROP INDEX IF EXISTS elysium.idx_layout_elements_layout_id;
DROP INDEX IF EXISTS elysium.idx_layout_editors_editor_id;
DROP INDEX IF EXISTS elysium.idx_layout_editors_layout_id;
DROP INDEX IF EXISTS elysium.idx_user_layouts_name;
DROP INDEX IF EXISTS elysium.idx_user_layouts_creator_id;

-- Удаление таблиц
DROP TABLE IF EXISTS elysium.layout_elements;
DROP TABLE IF EXISTS elysium.layout_editors;
DROP TABLE IF EXISTS elysium.user_layouts;
DROP TABLE IF EXISTS elysium.root_elements;
