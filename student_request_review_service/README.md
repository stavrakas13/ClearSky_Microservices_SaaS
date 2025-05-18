# STUDENT REQUESTS MICROSERVICE

## Grade Review Request

Students can make review requests for their grades.

Deployed with Docker + Rabbitmq + PostgresSQL with command: docker-compose up --build

## Functions + examples:

# 1. `student.postNewRequest`

Inserts new requests in reviewsdb.

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

---

# 2. `student.getRequestStatus`

Retuns details of a review.

Message Body:

```json
{
  "params": {
    "course_id": "101",
    "exam_period": "spring 2025",
    "user_id": "42"
  },
  "body": {

  }
}
```

---

# 3. `student.updateInstructorResponse`

Update reviewsdb with instructor responses.

Message Body:

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