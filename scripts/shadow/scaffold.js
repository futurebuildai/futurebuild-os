#!/usr/bin/env node
/**
 * Shadow Protocol: Scaffold Script
 * See specs/STEP63_specs.md
 *
 * Generates missing shadow documentation files for source code.
 * CONSTRAINT: Never reads source file content (Zero Trust Security).
 */

import { readdir, stat, mkdir, writeFile, access } from 'fs/promises';
import { join, dirname, basename, extname, relative } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const PROJECT_ROOT = join(__dirname, '..', '..');

// Configuration: Source Root -> Shadow Root
// Note: Using 'backend/shadow' instead of 'server/shadow' because 'server' is a binary file
const SOURCE_SHADOW_MAP = [
    { source: 'frontend/src', shadow: 'frontend/shadow' },
    { source: 'internal', shadow: 'backend/shadow/internal' },
    { source: 'cmd', shadow: 'backend/shadow/cmd' },
    { source: 'pkg', shadow: 'backend/shadow/pkg' },
];

// Inclusion: File extensions to process
const INCLUDED_EXTENSIONS = new Set(['.ts', '.tsx', '.go', '.js', '.py']);

// Exclusion patterns
const EXCLUDED_DIRS = new Set(['node_modules', '.git', 'assets', 'fixtures', 'mocks']);
const TEST_FILE_PATTERNS = [
    /\.test\.ts$/,
    /\.spec\.ts$/,
    /\.test\.tsx$/,
    /\.spec\.tsx$/,
    /_test\.go$/,
    /\.test\.js$/,
    /\.spec\.js$/,
];
const EXCLUDED_EXTENSIONS = new Set(['.css', '.scss', '.png', '.svg', '.jpg', '.jpeg', '.gif', '.ico']);

// L7 Template for shadow files
function getL7Template(filename) {
    return `# ${filename}

## Intent
*   **High Level:** [Auto-filled: Pending documentation]
*   **Business Value:** [Auto-filled: Pending documentation]

## Responsibility
*   State what this component handles.

## Key Logic
*   Describe flows and state management.

## Dependencies
*   **Upstream:** [Incoming calls]
*   **Downstream:** [Outgoing calls]
`;
}

// Check if a file exists
async function fileExists(filePath) {
    try {
        await access(filePath);
        return true;
    } catch {
        return false;
    }
}

// Check if directory exists
async function dirExists(dirPath) {
    try {
        const stats = await stat(dirPath);
        return stats.isDirectory();
    } catch {
        return false;
    }
}

// Check if file should be included
function shouldIncludeFile(filename) {
    const ext = extname(filename).toLowerCase();

    // Check extension inclusion
    if (!INCLUDED_EXTENSIONS.has(ext)) {
        return false;
    }

    // Check excluded extensions (redundant but explicit)
    if (EXCLUDED_EXTENSIONS.has(ext)) {
        return false;
    }

    // Check test file patterns
    for (const pattern of TEST_FILE_PATTERNS) {
        if (pattern.test(filename)) {
            return false;
        }
    }

    return true;
}

// Check if directory should be traversed
function shouldTraverseDir(dirname) {
    return !EXCLUDED_DIRS.has(dirname);
}

// Recursively walk directory and collect valid source files
async function walkDirectory(dirPath, baseDir, files = []) {
    let entries;
    try {
        entries = await readdir(dirPath, { withFileTypes: true });
    } catch (err) {
        // Directory doesn't exist or can't be read
        return files;
    }

    for (const entry of entries) {
        const fullPath = join(dirPath, entry.name);

        if (entry.isDirectory()) {
            if (shouldTraverseDir(entry.name)) {
                await walkDirectory(fullPath, baseDir, files);
            }
        } else if (entry.isFile()) {
            if (shouldIncludeFile(entry.name)) {
                const relativePath = relative(baseDir, fullPath);
                files.push({ fullPath, relativePath });
            }
        }
    }

    return files;
}

// Get shadow file path from source file
function getShadowPath(relativePath, shadowRoot) {
    const ext = extname(relativePath);
    const withoutExt = relativePath.slice(0, -ext.length);
    return join(PROJECT_ROOT, shadowRoot, withoutExt + '.md');
}

// Track created shadow files to detect collisions
const createdShadowFiles = new Map();

// Main scaffold function
async function scaffold() {
    console.log('Shadow Protocol: Scaffolding...\n');

    let totalCreated = 0;
    let totalSkipped = 0;
    let collisions = 0;

    for (const { source, shadow } of SOURCE_SHADOW_MAP) {
        const sourceDir = join(PROJECT_ROOT, source);

        // Check if source directory exists
        if (!(await dirExists(sourceDir))) {
            console.log(`Skipping ${source} (directory does not exist)`);
            continue;
        }

        console.log(`Processing: ${source} -> ${shadow}`);

        const files = await walkDirectory(sourceDir, sourceDir);

        for (const { relativePath } of files) {
            const shadowPath = getShadowPath(relativePath, shadow);
            const shadowDir = dirname(shadowPath);

            // Check for collision (multiple source files -> same shadow file)
            if (createdShadowFiles.has(shadowPath)) {
                const existingSource = createdShadowFiles.get(shadowPath);
                console.log(`  WARNING: Collision detected:`);
                console.log(`    - ${existingSource}`);
                console.log(`    - ${join(source, relativePath)}`);
                console.log(`    -> ${relative(PROJECT_ROOT, shadowPath)}`);
                collisions++;
                continue;
            }

            // Check if shadow file already exists
            if (await fileExists(shadowPath)) {
                totalSkipped++;
                continue;
            }

            // Create parent directories (mkdir -p equivalent)
            await mkdir(shadowDir, { recursive: true });

            // Get filename for template
            const filename = basename(relativePath, extname(relativePath));

            // Write shadow file with L7 template
            await writeFile(shadowPath, getL7Template(filename), 'utf8');

            // Track for collision detection
            createdShadowFiles.set(shadowPath, join(source, relativePath));

            console.log(`  Created: ${relative(PROJECT_ROOT, shadowPath)}`);
            totalCreated++;
        }
    }

    console.log('\n--- Shadow Protocol Scaffold Complete ---');
    console.log(`Created: ${totalCreated} files`);
    console.log(`Skipped: ${totalSkipped} files (already exist)`);
    if (collisions > 0) {
        console.log(`Collisions: ${collisions} (warnings logged above)`);
    }
}

// Run
scaffold().catch(err => {
    console.error('Shadow scaffold failed:', err);
    process.exit(1);
});
