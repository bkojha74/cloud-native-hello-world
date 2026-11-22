db = db.getSiblingDB("hello_world");

db.messages.updateOne(
  { _id: "1" },
  { $setOnInsert: { _id: "1", message: "Hello World" } },
  { upsert: true }
);

print("MongoDB seed data initialized successfully.");
