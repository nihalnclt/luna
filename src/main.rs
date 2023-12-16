use serde::{Deserialize, Serialize};
use std::{
    collections::HashMap,
    env,
    fs::File,
    io::{self, Write},
};

// #[allow(unused)]
#[derive(Serialize, Deserialize, Debug)]
struct PackageJson {
    name: String,
    version: String,
    description: String,
    main: String,
    scripts: HashMap<String, String>,
    keywords: Vec<String>,
    author: String,
    license: String,
    // dependencies: HashMap<String, String>,
    // devDependencies: HashMap<String, String>,
}

impl PackageJson {
    fn new(name: &str) -> PackageJson {
        PackageJson {
            name: String::from(name),
            version: String::from("1.0.0"),
            description: String::from(""),
            main: String::from("index.js"),
            scripts: HashMap::new(),
            keywords: Vec::new(),
            author: String::from(""),
            license: String::from("ISC"),
        }
    }

    fn write_to_file(&self) {
        let mut file = File::create("package.json").expect("Couldn't create file");
        let json_string = serde_json::to_string_pretty(&self).expect("Couldn't convert");

        file.write_all(json_string.as_bytes())
            .expect("Couldn't write to file");
    }
}

fn main() {
    let args: Vec<String> = env::args().collect();

    match args[1].as_str() {
        "init" => {
            let current_dir = env::current_dir().unwrap();
            let cwd_name: std::borrow::Cow<'_, str> = current_dir
                .file_name()
                .unwrap_or_default()
                .to_string_lossy();

            let mut package_json = PackageJson::new(&cwd_name);

            println!("This will walk you through creating a package.json file.");
            println!("It only covers the most common items, and tries to guess sensible defaults.");

            package_json.scripts.insert(
                "test".to_string(),
                "echo \"Error: no test specified\" && exit 1".to_string(),
            );
            package_json.name =
                read_input(&format!("package name: ({}) ", package_json.name)).unwrap();
            package_json.version =
                read_input(&format!("version: ({}) ", package_json.version)).unwrap();
            package_json.description = read_input("description: ").unwrap();
            package_json.author = read_input("author: ").unwrap();
            package_json.license =
                read_input(&format!("license: ({}) ", package_json.license)).unwrap();

            package_json.write_to_file();

            dbg!(package_json);
        }
        _ => {}
    }
    println!("Hello, world!{:?}", args);
}

fn read_input_default(prompt: &str, default: &str) -> String {
    print!("{}", prompt);
    io::stdout().flush().unwrap();

    let mut input = String::new();
    io::stdin().read_line(&mut input).unwrap();

    input.trim().to_string()
}
