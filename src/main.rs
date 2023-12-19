use flate2::read::GzDecoder;
use node_semver::{Range, Version};
use reqwest::blocking::get;
use serde::{Deserialize, Serialize};
use std::{
    collections::{HashMap, HashSet},
    env,
    fs::{self, File},
    io::{self, Write},
    path::Path,
    time::Instant,
};
use tar::Archive;

// #[allow(unused)]
#[derive(Serialize, Deserialize, Debug)]
#[serde(rename_all = "camelCase")]
struct PackageJson {
    name: String,
    version: String,
    description: String,
    main: String,
    scripts: HashMap<String, String>,
    keywords: Vec<String>,
    author: String,
    license: String,
    dependencies: Option<HashMap<String, String>>,
    dev_dependencies: Option<HashMap<String, String>>,
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
            dependencies: Some(HashMap::new()),
            dev_dependencies: Some(HashMap::new()),
        }
    }

    fn write_to_file(&self) {
        let mut file = File::create("package.json").expect("Couldn't create file");
        let json_string = serde_json::to_string_pretty(&self).unwrap();

        file.write_all(json_string.as_bytes())
            .expect("Couldn't write to file");
    }
}

fn download_tgz(url: &str, destination: &str) -> Result<(), Box<dyn std::error::Error>> {
    if let Ok(response) = get(url) {
        println!("{}", response.status());
        if response.status().is_success() {
            let response_bytes = response.bytes().unwrap();
            let gz_decoder = GzDecoder::new(response_bytes.as_ref());
            let mut archive = Archive::new(gz_decoder);

            // fs::create_dir_all(destination).unwrap();
            // println!("destination is: {}", destination);
            // archive.unpack(destination).unwrap();

            for entry in archive.entries().unwrap() {
                let mut entry = entry.unwrap();
                let path = entry.path().unwrap();

                let full_path =
                    Path::new(destination).join(path.strip_prefix("package/").unwrap_or(&path));

                if let Some(parent) = full_path.parent() {
                    fs::create_dir_all(parent).unwrap();
                }

                entry.unpack(full_path).unwrap();
            }
        }
    }

    Ok(())
}

#[derive(Serialize, Deserialize, Debug)]
struct Signature {
    keyid: String,
    sig: String,
}

#[derive(Serialize, Deserialize, Debug)]
#[serde(rename_all = "camelCase")]
struct PackageDist {
    integrity: String,
    shasum: String,
    tarball: String,
    file_count: usize,
    unpacked_size: usize,
    signatures: Vec<Signature>,
}

#[derive(Serialize, Deserialize, Debug)]
struct PackageVersion {
    name: String,
    version: String,
    dist: PackageDist,
    dependencies: Option<HashMap<String, String>>,
}

#[derive(Serialize, Deserialize, Debug)]
struct PackageInfo {
    name: String,
    #[serde(rename = "dist-tags")]
    dist_tags: HashMap<String, String>,
    versions: HashMap<String, PackageVersion>,
    license: String,
}

fn fetch_package_info() -> PackageInfo {
    let url = String::from("https://registry.npmjs.org/@webdiari/common");
    let response = get(&url).unwrap();
    let json_string = response.text().unwrap();
    let package_info: PackageInfo = serde_json::from_str(&json_string).unwrap();
    package_info
}

#[derive(Debug, Eq, Hash, PartialEq)]
struct DepReq {
    name: String,
}

#[derive(Debug, Eq, PartialEq)]
struct DependencyGraph {
    pub relations: HashMap<DepReq, String>,
}

impl DependencyGraph {
    fn new() -> Self {
        DependencyGraph {
            relations: HashMap::new(),
        }
    }

    fn add_node(&mut self, dep: DepReq) {
        self.relations.insert(dep, String::from("value"));
    }

    fn add_edge(&mut self, source: DepReq, target: DepReq) {
        self.relations
            .entry(source)
            .or_insert_with(Vec::new)
            .push(target);
    }
}

// if i created a graph then i have non-duplicated dependencies with connections
// for installing loop through all and download it's dependencies
// 

fn main() {
    let args: Vec<String> = env::args().collect();
    let green_code = "\x1b[32m";
    let reset_code = "\x1b[0m";

    match args[1].as_str() {
        "init" => {
            let start_time = Instant::now();
            let current_dir = env::current_dir().unwrap();
            let cwd_name: std::borrow::Cow<'_, str> = current_dir
                .file_name()
                .unwrap_or_default()
                .to_string_lossy();

            let mut package_json = PackageJson::new(&cwd_name);
            package_json.scripts.insert(
                "test".to_string(),
                "echo \"Error: no test specified\" && exit 1".to_string(),
            );

            if args.len() > 2 && args[2] == "-y" {
                package_json.write_to_file();
            } else {
                println!("This will walk you through creating a package.json file.");
                println!(
                    "It only covers the most common items, and tries to guess sensible defaults."
                );
                println!("");

                package_json.name = read_input_default("package name", &package_json.name).unwrap();
                package_json.version =
                    read_input_default("version", &package_json.version).unwrap();
                package_json.description = read_input("description: ").unwrap();
                package_json.author = read_input("author: ").unwrap();
                package_json.license =
                    read_input_default("license", &package_json.license).unwrap();
            }

            package_json.write_to_file();

            let end_time = Instant::now();

            println!("");
            println!(
                "{}success{} saved package.json file",
                green_code, reset_code
            );
            println!(
                "Done in {:.2}s",
                end_time.duration_since(start_time).as_secs_f64()
            );
        }
        "install" | "i" => {
            // let packages: Vec<&String> = args[2..].iter().collect();

            let package_info = fetch_package_info();
            let version = "~0.0.2";

            let dependencies_tree: HashMap<String, PackageVersion>;
            // depName, dependencyDetails

            // dep and it's dependencies
            // name: {version, ...etc}
            // if name then check same version
            // if not add that to tree
            // lastly loop tree and download dependencies if not in cache
            // before adding to top level check is there two availalable
            // if there two with different versions
            // webdiari 0.1 installed on top
            // then when i install 0.3 then 0.1 need to move it's specific dependencies
            // save paths on each package = ["node_modules/@webdiari/node_modules"]

            // mime-types@~2.1.24:
            //  version "2.1.35"
            //  resolved "https://registry.npmjs.org/mime-types/-/mime-types-2.1.35.tgz"
            //  integrity sha512-ZDY+bPm5zTTF+YpCrAU9nK0UgICYPT0QtT1NZWFv4s++TNkcgVaT0g6+4R2uI4MjQjzysHB1zxuWL50hzaeXiw==
            //  dependencies:
            //      mime-db "1.52.0"

            let dependencies: HashMap<String, String>;
            let dev_dependencies: HashMap<String, String>;

            // only for matching dependencies already downloaded
            // struct LockFile

            let temp: HashMap<String, Tree>;
            struct Tree {
                version: String,
                resolved: String,
                integrity: String,
                dependencies: HashMap<String, Tree>,
            }

            // loop through tree and download all dependencies but what about same packages

            let parsed_version: Range = version.parse().unwrap();
            let matching_versions: Vec<&String> = package_info
                .versions
                .keys()
                .filter(|&v| {
                    let version: Version = v.parse().unwrap();
                    version.satisfies(&parsed_version)
                })
                .collect();

            println!("Matching version {:?}", matching_versions);

            if let Some(&max_version_str) = matching_versions.iter().max() {
                println!(
                    "Version is {:?}",
                    package_info.versions.get(max_version_str)
                );
            } else {
                println!("Version Not Found");
            }

            // if let Some(package_version) = package_info.versions.get("0.0.7") {
            //     println!("package version: {:?}", package_version);
            //     if let Some(dependencies) = &package_version.dependencies {
            //         for (package, version) in dependencies {
            //             println!("Dependency: {:?}", package);
            //             println!("Version: {:?}", version);
            //         }
            //     }
            // }
            // let dependencies = dependency.get("dependencies").unwrap();
            // let package_info = fetch_package_info();

            let parts: Vec<&str> = package_info.name.split('/').collect();

            // TODO:
            // test this condition will work on every cases
            let mut org_name = String::from("");
            let mut package_name = package_info.name.as_str();
            if parts.len() >= 2 && parts[0].starts_with("@") {
                org_name = format!("/{}", parts[0]);
                package_name = parts[1];
            }

            let version;
            match package_info.dist_tags.get("latest") {
                Some(value) => version = value,
                None => panic!("couldn't find an version"),
            }

            let download_url = format!(
                "https://registry.npmjs.org/{}/-/{}-{}.tgz",
                package_info.name, package_name, version
            );

            println!("download url is {}", download_url);

            let destination = format!("node_modules{}/{}", org_name, package_name);
            download_tgz(&download_url, &destination);

            println!("Installing packages...{}{}", org_name, package_name);
        }
        _ => {
            println!("Unknown command: {}", args[1]);
            println!("");
            println!("To see a list of supported commands, run:");
            println!("nnpm help");
        }
    }
}

fn read_input(prompt: &str) -> io::Result<String> {
    print!("{}", prompt);
    io::stdout().flush()?;

    let mut input: String = String::new();
    io::stdin().read_line(&mut input)?;

    Ok(input.trim().to_string())
}

fn read_input_default(prompt: &str, default: &str) -> io::Result<String> {
    print!("{}: ({}) ", prompt, default);
    io::stdout().flush()?;

    let mut input: String = String::new();
    io::stdin().read_line(&mut input)?;

    if input.trim().is_empty() {
        input = default.to_string();

        return Ok(input);
    }

    Ok(input.trim().to_string())
}
