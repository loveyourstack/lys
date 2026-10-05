package lyspgdb

import (
	"fmt"
	"io/fs"
	"path"
	"strings"
)

// EnhanceMigrationFile enhances a migration file by replacing object placeholders with their corresponding DDL content from ddlAssets.
func EnhanceMigrationFile(fileName string, ddlAssets, migrationAssets fs.FS) (res string, err error) {

	// get migration file content
	res, err = getMigrationFile(fileName, migrationAssets)
	if err != nil {
		return "", err
	}

	const objectReplacementPrefix = "-- + "

	// read each line of content
	lines := strings.Split(res, "\n")
	for lineIndex, line := range lines {

		// skip lines that don't begin with the object replacement comment
		if !strings.HasPrefix(line, objectReplacementPrefix) {
			continue
		}

		// get the object name from the comment
		objName := strings.TrimPrefix(line, objectReplacementPrefix)
		objName = strings.TrimSuffix(objName, ";")

		// get the replacement content for the object
		replacement, err := getMigrationObjectReplacement(objName, ddlAssets)
		if err != nil {
			return "", err
		}

		// replace the line with the replacement content
		lines[lineIndex] = replacement
	}

	// rebuild content and add tx prefix and suffix
	res = fmt.Sprintf("BEGIN;\n%s\nROLLBACK;\n--COMMIT;", strings.Join(lines, "\n"))

	return res, nil
}

func getMigrationFile(fileName string, migrationAssets fs.FS) (content string, err error) {

	// try to find fileName in migrationAssets
	paths := []string{}
	err = fs.WalkDir(migrationAssets, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("unknown file err: %w", err)
		}
		if !d.IsDir() && d.Name() == fileName {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("fs.WalkDir failed: %w", err)
	}

	// ensure exactly one match was found
	if len(paths) == 0 {
		return "", fmt.Errorf("no match found for %s", fileName)
	}
	if len(paths) > 1 {
		return "", fmt.Errorf("expected exactly one match for %s, found %d", fileName, len(paths))
	}
	filePath := paths[0]

	// read contents of migration file
	contentB, err := fs.ReadFile(migrationAssets, filePath)
	if err != nil {
		return "", fmt.Errorf("fs.ReadFile failed: %w", err)
	}

	return string(contentB), nil
}

// getMigrationObjectReplacement retrieves the replacement content for a given object name from the ddlAssets.
// It expects ddlAssets to contain directories of schema names, each containing /funcs and /views subdirectories.
func getMigrationObjectReplacement(objName string, ddlAssets fs.FS) (replacement string, err error) {

	// objName should consist of schema and object name, separated by a dot (e.g., "public.my_table")
	parts := strings.SplitN(objName, ".", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid object name: %s", objName)
	}
	schemaName := parts[0]
	objectName := parts[1]

	// determine schema subdirectory based on object name prefix
	subdir := ""
	switch {
	case strings.HasPrefix(objectName, "f_"),
		strings.HasPrefix(objectName, "p_"),
		strings.HasPrefix(objectName, "tf_"):
		subdir = "funcs"
	case strings.HasPrefix(objectName, "mv_"),
		strings.HasPrefix(objectName, "v_"):
		subdir = "views"
	default:
		return "", fmt.Errorf("unsupported object name: %s", objectName)
	}

	// validate schema and object names to prevent directory traversal
	if strings.ContainsAny(schemaName, "/\\") || strings.ContainsAny(objectName, "/\\") {
		return "", fmt.Errorf("invalid object name: %s", objName)
	}

	// construct the path to the object file within ddlAssets
	objPath := path.Join(schemaName, subdir, objectName+".sql")

	// read the content of the object file
	contentB, err := fs.ReadFile(ddlAssets, objPath)
	if err != nil {
		return "", fmt.Errorf("fs.ReadFile failed for %s: %w", objPath, err)
	}

	return string(contentB), nil
}
