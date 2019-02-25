#![feature(proc_macro_hygiene, decl_macro, result_map_or_else)]

mod endpoints;
mod database;

#[macro_use]
extern crate rocket;
#[macro_use]
extern crate rocket_contrib;

extern crate mongodb;
extern crate bson;
extern crate chrono;

extern crate serde;
extern crate serde_json;
#[macro_use]
extern crate serde_derive;
extern crate argon2;

use std::sync::Mutex;

pub type DbConn<'cfg> = Mutex<crate::database::handler::DatabaseHandler<'cfg>>;

fn main() {
	let mut connection = database::handler::DatabaseHandler::new("192.168.99.100", 32769, "123123123123", "123123123123");
	match connection.connect("cactus", "c", "c") {
		Ok(()) => println!("Connected!"),
		Err(e) => println!("Error: {:?}", e)
	};

    rocket::ignite()
    	.manage(Mutex::new(connection))
	    .mount("/channel", routes! [
	    	endpoints::channel::get_channel, endpoints::channel::create_channel,
	    	endpoints::channel::get_command, endpoints::channel::delete_command,
	    	endpoints::channel::get_commands, endpoints::channel::create_command
	    ])
	    .mount("/state", routes! [
	    	endpoints::state::get_channel_state, endpoints::state::get_channel_service_state
	    ])
	    .mount("/auth", routes! [
	    	endpoints::authorization::get_service_auth, endpoints::authorization::update_service_auth
	    ])
	    .mount("/quote", routes! [
	    	endpoints::quote::get_quote, endpoints::quote::get_random_quote,
	    	endpoints::quote::get_quote_by_id, endpoints::quote::create_quote,
	    	endpoints::quote::delete_quote
	    ])
	    .launch();
}
