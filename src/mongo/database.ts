
import { Config } from "../config";

import * as Mongo from "mongodb";

export class MongoHandler {

	private connection: Mongo.Db;
	private commands: Mongo.Collection;
	private quotes: Mongo.Collection;

	constructor(private config: Config) {
	}

	public async connect() {
		const connectionURL = `mongodb://${this.config.mongo.username}:${this.config.mongo.password}@${this.config.mongo.host}:${this.config.mongo.port}/${this.config.mongo.database}`;
		Mongo.MongoClient.connect(connectionURL, {
				authSource: this.config.mongo.authdb
			},
			(err, db) => {
			if (err) {
				console.error(err);
				return;
			}
			this.connection = db;

			this.quotes = this.connection.collection("quotes");
			this.commands = this.connection.collection("commands");
		});
	}

	public async createQuote(quote: Quote) {
		const recent = await this.quotes.find({ channel: quote.channel }).sort({ quoteId: -1 }).limit(1).toArray();
		const quoteId = recent.length == 0 ? 1 : recent[0].quoteId + 1
		quote.quoteId = quoteId;

		this.quotes.insertOne(quote);
	}

	public async getAllQuotes(channel: string): Promise<Quote[]> {
		return await this.quotes.find({ channel }).toArray();
	}

	public async getQuote(channel: string, random: boolean, quoteId: number): Promise<Quote> {
		if (!quoteId && random) {
			const quotes = await this.quotes.aggregate([
				{
					"$sample": { "size": 1 }
				},
				{
					"$match": { channel }
				}
			]).toArray();
			if (quotes.length === 0) {
				return null;
			}
			return quotes[0];
		}
		if (quoteId === -1) {
			return null;
		}
		return await this.quotes.findOne({ channel, quoteId });
	}

	public async createCommand(command: Command) {
		this.commands.insertOne(command);
	}
}