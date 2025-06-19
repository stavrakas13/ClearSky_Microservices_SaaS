// app.js
import express from "express";
import session from "express-session";
import bodyParser from "body-parser";
import path from "path";
import { fileURLToPath } from "url";
import { createProxyMiddleware } from "http-proxy-middleware";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const app = express();

// 1) JSON + URL-encoded parsers (so fetch(..., { body: JSON.stringify(...) }) works)
app.use(bodyParser.json());
app.use(bodyParser.urlencoded({ extended: false }));

// 2) Session for your EJS pages
app.use(
  session({
    secret: "change-this-secret",
    resave: false,
    saveUninitialized: true,
  })
);

// 3) Expose user + currentUrl to your templates
app.use((req, res, next) => {
  res.locals.user       = req.session.user || null;
  res.locals.currentUrl = req.originalUrl;
  next();
});

// 4) Serve static assets (public/js, styles.css, etc)
app.use(express.static(path.join(__dirname, "public")));

// 5) Proxy *only* your API endpoints to the Go backend
const API_TARGET = process.env.GO_API_URL || "http://localhost:3001";
const apiPaths = [
  "/purchase",
  "/mycredits",
  "/spending",
  "/registration",
  "/upload_init",
  "/stats/persist",
  "/stats/distributions",
  "/personal/courses",
  "/personal/grades",
  "/student/reviewRequest",
  "/student/status",
  "/instructor/review-list",
  "/instructor/reply",
  "/user/register",
  "/user/login",
  "/user/delete",
  "/user/google-login",
];
app.use(
  apiPaths,
  createProxyMiddleware({
    target: API_TARGET,
    changeOrigin: true,
    logLevel: "warn",
  })
);

// 6) EJS view engine setup
app.set("view engine", "ejs");
app.set("views", path.join(__dirname, "views"));

// --- Your existing page routes follow: ---

// Public sign-up page
app.get("/signup", (req, res) =>
  res.render("institution/userManagement", {
    user: null,
    title: "Sign Up",
  })
);

// auth helper
function auth(role) {
  return (req, res, next) => {
    if (!req.session.user) return res.redirect("/login");
    if (role && req.session.user.role !== role)
      return res.redirect(`/${req.session.user.role}`);
    next();
  };
}

// mock users
const users = { alice: "student", bob: "instructor", iris: "institution" };

// root → role dashboard
app.get("/", (req, res) => {
  const u = req.session.user;
  if (!u) return res.redirect("/login");
  return res.redirect(`/${u.role}`);
});

// login
app.get("/login", (req, res) =>
  res.render("login", { title: "Log in", error: null, user: null })
);
app.post("/login", (req, res) => {
  const { username, password } = req.body;
  const role = users[username];
  if (!role || password !== "1234") {
    return res.render("login", {
      title: "Log in",
      error: "Invalid credentials",
      user: null,
    });
  }
  req.session.user = { username, role };
  return res.redirect(`/${role}`);
});
app.get("/logout", (req, res) =>
  req.session.destroy(() => res.redirect("/login"))
);

// student routes
app.get("/student", auth("student"), (r, s) =>
  s.render("student/dashboard", { user: r.session.user, title: "Dashboard" })
);
app.get("/student/statistics", auth("student"), (r, s) =>
  s.render("student/statistics", { user: r.session.user, title: "Statistics" })
);
app.get("/student/my-courses", auth("student"), (r, s) =>
  s.render("student/myCourses", { user: r.session.user, title: "My Courses" })
);
app.get("/student/request", auth("student"), (r, s) =>
  s.render("student/reviewRequest", { user: r.session.user, title: "Review Request" })
);
app.get("/student/status", auth("student"), (r, s) =>
  s.render("student/reviewStatus", { user: r.session.user, title: "Review Status" })
);
app.get("/student/personal", auth("student"), (r, s) =>
  s.render("student/personal", { user: r.session.user, title: "Personal Grades" })
);

// instructor routes
app.get("/instructor", auth("instructor"), (r, s) =>
  s.render("instructor/dashboard", { user: r.session.user, title: "Dashboard" })
);
app.get("/instructor/post-initial", auth("instructor"), (r, s) =>
  s.render("instructor/postInitial", { user: r.session.user, title: "Post Initial" })
);
app.get("/instructor/review-list", auth("instructor"), (r, s) =>
  s.render("instructor/reviewList", { user: r.session.user, title: "Review List" })
);
app.get("/instructor/reply", auth("instructor"), (r, s) =>
  s.render("instructor/replyForm", { user: r.session.user, title: "Reply Form" })
);
app.get("/instructor/post-final", auth("instructor"), (r, s) =>
  s.render("instructor/postFinal", { user: r.session.user, title: "Post Final" })
);
app.get("/instructor/statistics", auth("instructor"), (r, s) =>
  s.render("instructor/statistics", { user: r.session.user, title: "Statistics" })
);

// institution routes
app.get("/institution", auth("institution"), (r, s) =>
  s.render("institution/dashboard", { user: r.session.user, title: "Dashboard" })
);
app.get("/institution/register", auth("institution"), (r, s) =>
  s.render("institution/register", { user: r.session.user, title: "Register" })
);
app.get("/institution/purchase", auth("institution"), (r, s) =>
  s.render("institution/purchase", { user: r.session.user, title: "Purchase" })
);
app.get("/institution/user-management", auth("institution"), (r, s) =>
  s.render("institution/userManagement", { user: r.session.user, title: "Users" })
);
app.get("/institution/statistics", auth("institution"), (r, s) =>
  s.render("institution/statistics", { user: r.session.user, title: "Statistics" })
);

// start server
const PORT = process.env.PORT || 3000;
app.listen(PORT, () => console.log(`✔ Front-end server listening on http://localhost:${PORT}`));
