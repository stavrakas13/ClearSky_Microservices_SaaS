# INSTRUCTOR RESPONSE MICROSERVICE

## Reply to Review Request

Instructors reply to Grade Review Requests made by the students.

Deployed with Docker + Rabbitmq + PostgresSQL with command: docker-compose up --build

## Functions + examples:

# 1. `instructor.postResponse`  

Updates reviewdb with reply.

Input Message Body:

```json
{
  "params": {
    "course_id": "101",
    "exam_period": "spring 2025",
    "user_id": "42"
  },
  "body": {
    "instructor_reply_message": "We will consider your case.",
    "instructor_action": "Will be considered"
  }
}
```

---

# 2. `instructor.getRequestsList`

Returns list of all pending requests, given a cource id.

Input Message Body:

```json
{
  "params": {
    "course_id": "101"
  },
  "body": {}
}
```

---

# 3. `instructor.getRequestInfo`

Returns a review summary.

Message Body:

```json
{
  "params": {
    "course_id": "101",
    "exam_period": "spring 2025",
    "user_id": "42"
  },
  "body": {}
}
```

# 4. `instructor.insertStudentRequest`

Updates reviewdb with new requests.

Message Body:

```json
{
  "params": {
    "course_id": "101",
    "exam_period": "spring 2025"
  },
  "body": {
    "user_id": 42,
    "student_message": "Please review my grade."
  }
}
```