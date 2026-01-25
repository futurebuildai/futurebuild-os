#!/usr/bin/env node
/**
 * Shadow Protocol: Check Script (The Enforcer)
 * See specs/STEP63_specs.md
 *
 * Verifies every valid source file has a corresponding shadow file.
 * EXIT 1 if any are missing, EXIT 0 if all present.
 */

import { readdir, stat, access } from 'fs/promises';
import { join, dirname, extname, relative } from 'path';
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

// Main check function
async function check() {
    console.log('Shadow Protocol: Checking...\n');

    const missingFiles = [];
    let totalChecked = 0;

    // Track shadow paths to handle collisions (multiple sources -> same shadow)
    const checkedShadows = new Set();

    for (const { source, shadow } of SOURCE_SHADOW_MAP) {
        const sourceDir = join(PROJECT_ROOT, source);

        // Check if source directory exists
        if (!(await dirExists(sourceDir))) {
            continue;
        }

        const files = await walkDirectory(sourceDir, sourceDir);

        for (const { relativePath } of files) {
            const shadowPath = getShadowPath(relativePath, shadow);
            const sourceFile = join(source, relativePath);
            const shadowFile = relative(PROJECT_ROOT, shadowPath);

            // Skip if we've already checked this shadow path (collision handling)
            if (checkedShadows.has(shadowPath)) {
                continue;
            }
            checkedShadows.add(shadowPath);

            totalChecked++;

            if (!(await fileExists(shadowPath))) {
                missingFiles.push({
                    source: sourceFile,
                    shadow: shadowFile,
                });
            }
        }
    }

    console.log(`Checked: ${totalChecked} source files\n`);

    if (missingFiles.length > 0) {
        console.log('ERROR: Shadow Protocol Violation.');
        console.log('The following files are missing shadow docs:\n');

        for (const { source, shadow } of missingFiles) {
            console.log(`  Source: ${source}`);
            console.log(`  Shadow: ${shadow}\n`);
        }

        console.log(`Total missing: ${missingFiles.length}`);
        console.log("\nRun 'npm run shadow:scaffold' to fix this.");
        process.exitCode = 1;
        return;
    }

    console.log('Shadow Protocol Verified.');
    console.log(`All ${totalChecked} source files have shadow documentation.`);
    process.exitCode = 0;
}

// Run
check()
    .then(() => {
        // Ensure exit code is set (default to 0 if not already set)
        if (process.exitCode === undefined) {
            process.exitCode = 0;
        }
    })
    .catch(err => {
        console.error('Shadow check failed:', err);
        process.exitCode = 1;
    });
