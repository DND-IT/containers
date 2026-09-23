// Prints `export NAME='value'` lines for the entrypoint to eval: the strings under
// SECRET_ENV_KEY in the JSON secret SECRET_ENV_ARN. Exits non-zero, printing nothing,
// when the secret cannot be read, has no such key, or holds anything but strings,
// so the server never starts half-configured.
import {
  GetSecretValueCommand,
  SecretsManagerClient,
} from "@aws-sdk/client-secrets-manager";

const fail = (message) => {
  console.error(`secret-env: ${message}`);
  process.exit(1);
};

const arn = process.env.SECRET_ENV_ARN;
const key = process.env.SECRET_ENV_KEY;
if (!key) fail("SECRET_ENV_ARN is set but SECRET_ENV_KEY is not");

const region = arn.split(":")[3];
if (!arn.startsWith("arn:") || !region) fail(`SECRET_ENV_ARN is not a secret ARN: ${arn}`);

let secret;
try {
  const client = new SecretsManagerClient({ region });
  const { SecretString } = await client.send(new GetSecretValueCommand({ SecretId: arn }));
  secret = JSON.parse(SecretString ?? "");
} catch (error) {
  fail(`cannot read ${arn}: ${error.name}: ${error.message}`);
}

const entry = secret?.[key];
if (entry === null || typeof entry !== "object" || Array.isArray(entry)) {
  fail(`${arn} has no object under "${key}"`);
}

const quote = (value) => `'${value.replaceAll("'", `'\\''`)}'`;
const lines = [];
for (const [name, value] of Object.entries(entry)) {
  if (!/^[A-Za-z_][A-Za-z0-9_]*$/.test(name)) fail(`"${key}.${name}" is not a variable name`);
  if (typeof value !== "string") fail(`"${key}.${name}" is not a string`);
  lines.push(`export ${name}=${quote(value)}`);
}
process.stdout.write(lines.join("\n") + "\n");
