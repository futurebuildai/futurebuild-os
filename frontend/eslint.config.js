import eslint from '@eslint/js';
import tseslint from 'typescript-eslint';
import litPlugin from 'eslint-plugin-lit';
import wcPlugin from 'eslint-plugin-wc';

// See FRONTEND_SCOPE.md Section 13 - Standards & Linting
// Using ESLint 9.x flat config format with Lit-specific rules
export default tseslint.config(
    eslint.configs.recommended,
    ...tseslint.configs.strictTypeChecked,
    litPlugin.configs['flat/recommended'],
    wcPlugin.configs['flat/recommended'],
    {
        languageOptions: {
            parserOptions: {
                projectService: true,
                tsconfigRootDir: import.meta.dirname,
            },
        },
        rules: {
            // TypeScript Strictness
            '@typescript-eslint/no-explicit-any': 'error',
            '@typescript-eslint/explicit-function-return-type': 'warn',
            '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],

            // Lit Best Practices
            'lit/no-legacy-template-syntax': 'error',
            'lit/attribute-names': 'error',
            'lit/binding-positions': 'error',
            'lit/no-invalid-html': 'error',
            'lit/no-useless-template-literals': 'warn',

            // Web Components
            'wc/no-constructor-attributes': 'error',
            'wc/no-invalid-element-name': 'error',
        },
    },
    {
        // Config files don't need strict type checking
        files: ['*.config.js', '*.config.ts'],
        ...tseslint.configs.disableTypeChecked,
    },
    {
        ignores: ['dist/**', 'node_modules/**'],
    }
);
