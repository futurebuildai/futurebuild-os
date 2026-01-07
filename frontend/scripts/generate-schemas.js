import { resolve } from "path";
import { writeFileSync, mkdirSync } from "fs";
import * as TJS from "ts-json-schema-generator";

const config = {
    path: resolve("src/types/models.ts"),
    tsconfig: resolve("tsconfig.json"),
    type: "*", // Generate all types
};

const outputDir = resolve("src/types/schemas");
mkdirSync(outputDir, { recursive: true });

const generator = TJS.createGenerator(config);
const types = ["Forecast", "Contact", "InvoiceExtraction", "GanttData"];

types.forEach((type) => {
    const schema = generator.createSchema(type);
    const schemaString = JSON.stringify(schema, null, 2);
    writeFileSync(resolve(outputDir, `${type}.schema.json`), schemaString);
    console.log(`Generated schema for ${type}`);
});
