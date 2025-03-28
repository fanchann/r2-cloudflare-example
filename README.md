# R2 cloudflare example

# API documentation
## **1. Upload File**
**URL:** `http://localhost:8080/upload`

**Method:** `POST`

**Body Request:** `multipart/form-data`

**Description:** Upload file.

### **Curl Command:**
```sh
curl -X POST "http://localhost:8080/upload" \
     -F "file=@/path/to/your/file.jpg"
```

---

## **2. Get List of Files**
**URL:** `http://localhost:8080/lists`

**Method:** `GET`

**Body Request:** `null`

**Description:** Get list file from server.

### **Curl Command:**
```sh
curl -X GET "http://localhost:8080/lists"
```

---

## **3. Make File Public**
**URL:** `http://localhost:8080/public`

**Method:** `POST`

**Body Request:** `JSON`

**Description:** Set status file from private to public.

### **Body Request Format:**
```json
{
  "file_id": "your_file_id",
  "duration":"duration"
}
```

### **Curl Command:**
```sh
curl -X POST "http://localhost:8080/public" \
     -H "Content-Type: application/json" \
     -d '{"file_id": "your_file_id","duration":"duration"}'
```

---

## **4. Get File by ID**
**URL:** `http://localhost:8080/file/:id`

**Method:** `GET`

**Body Request:** `null`

**Description:** Get file by id.

### **Curl Command:**
```sh
curl -X GET "http://localhost:8080/file/your_file_id"
```

