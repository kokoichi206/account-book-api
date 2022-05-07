# ER 図

``` mermaid
erDiagram

users ||--o{ expenses: "go to"

users {
	bigint id PK
	string name
	string password
	string email 
	int age
	bigint balance
	timestamp password_changed_at
	timestamp created_at
}

sessions }|--||users : "have"
sessions {
	uuid id PK
	string user_id FK
	string resresh_token
	string user_agent
	string client_ip
	timestamp created_at
	timestamp expires_at
}

expenses {
	bigint id PK
	bigint user_id FK
	bigint category_id FK
	bigint amount
	bigint food_receipt_id FK
	string comment
	timestamp created_at
}

expenses }o--||categories : "belong to"
categories {
	bigint id PK
	string name
}

food_receipts |o--||expenses : "may have"
food_receipts {
	bigint id PK
	string store_name
}

food_receipts ||--|{food_receipt_contents : ""
food_receipt_contents {
	bigint id PK
	bigint food_receipt_id FK
	bigint food_content_id FK
	bigint amount
}

food_receipt_contents }o--||food_contents : ""
food_contents {
	bigint id PK
	string name
	string calories
	float4 lipid
	float4 carbohydrate
	float4 Protein
}

transfers }o--||users : do
transfers {
	bigint id PK
	bigint from_user_id FK
	bigint to_user_id FK
	bigint amount
	timestamp created_at
}
```
