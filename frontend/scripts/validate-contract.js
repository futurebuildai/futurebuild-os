import Ajv from "ajv";
import { readFileSync, readdirSync } from "fs";
import { resolve } from "path";

const ajv = new Ajv({ allErrors: true });
const samplesDir = resolve("../internal/contract_validation/samples");
const schemasDir = resolve("src/types/schemas");

const files = readdirSync(samplesDir).filter(f => f.endsWith(".json"));

let exitCode = 0;

files.forEach(file => {
    const typeName = file.replace(".json", "");
    const sample = JSON.parse(readFileSync(resolve(samplesDir, file), "utf-8"));
    const schema = JSON.parse(readFileSync(resolve(schemasDir, `${typeName}.schema.json`), "utf-8"));

    const validate = ajv.compile(schema);
    const valid = validate(sample);

    if (valid) {
        console.log(`✅ ${typeName}: Contract Validated`);
    } else {
        console.error(`❌ ${typeName}: Contract Violation`);
        console.error(ajv.errorsText(validate.errors));
        exitCode = 1;
    }
});

process.exit(exitCode);
