import express from "express";
import session from "express-session";
import bodyParser from "body-parser";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const app = express();

/* ---------- view engine ---------- */
app.set("view engine", "ejs");
app.set("views", path.join(__dirname, "views"));

/* ---------- middleware ---------- */
app.use(express.static(path.join(__dirname, "public")));
app.use(bodyParser.urlencoded({ extended: false }));
app.use(
  session({
    secret: "change‑this‑secret",
    resave: false,
    saveUninitialized: true
  })
);

/* ---------- auth helper ---------- */
function auth(role) {
  return (req, res, next) => {
    if (!req.session.user) return res.redirect("/login");
    if (role && req.session.user.role !== role)
      return res.redirect(`/${req.session.user.role}`);
    next();
  };
}

/* ---------- mock users ---------- */
const users = { alice: "student", bob: "instructor", iris: "institution" };

/* ---------- routes ---------- */
app.get("/", (req, res) => res.redirect("/login"));

app.get("/login", (req, res) =>
  res.render("login", { title: "Log in", error: null, user: null })
);

app.post("/login", (req, res) => {
  const { username, password } = req.body;
  const role = users[username];
  if (!role || password !== "1234")
    return res.render("login", { title: "Log in", error: "Invalid credentials", user: null });
  req.session.user = { username, role };
  res.redirect(`/${role}`);
});
app.get("/logout", (req, res) => req.session.destroy(() => res.redirect("/login")));

/* ---- student ---- */
app.get("/student",           auth("student"), (r,s)=>s.redirect("/student/statistics"));
app.get("/student/statistics",auth("student"), (r,s)=>s.render("student/statistics",{title:"Course stats",user:r.session.user}));
app.get("/student/my-courses",auth("student"), (r,s)=>s.render("student/myCourses",{title:"My courses",user:r.session.user}));
app.get("/student/request",   auth("student"), (r,s)=>s.render("student/reviewRequest",{title:"Grade review request",user:r.session.user}));
app.get("/student/status",    auth("student"), (r,s)=>s.render("student/reviewStatus",{title:"Review status",user:r.session.user}));

/* ---- instructor ---- */
app.get("/instructor",               auth("instructor"),(r,s)=>s.redirect("/instructor/statistics"));
app.get("/instructor/statistics",    auth("instructor"),(r,s)=>s.render("instructor/statistics",{title:"Course stats",user:r.session.user}));
app.get("/instructor/post-initial",  auth("instructor"),(r,s)=>s.render("instructor/postInitial",{title:"Post initial grades",user:r.session.user}));
app.get("/instructor/post-final",    auth("instructor"),(r,s)=>s.render("instructor/postFinal",{title:"Post final grades",user:r.session.user}));
app.get("/instructor/review-list",   auth("instructor"),(r,s)=>s.render("instructor/reviewList",{title:"Reply list",user:r.session.user}));
app.get("/instructor/reply",         auth("instructor"),(r,s)=>s.render("instructor/replyForm",{title:"Reply form",user:r.session.user}));

/* ---- institution ---- */
app.get("/institution",                auth("institution"),(r,s)=>s.redirect("/institution/statistics"));
app.get("/institution/statistics",     auth("institution"),(r,s)=>s.render("institution/statistics",{title:"Course stats",user:r.session.user}));
app.get("/institution/register",       auth("institution"),(r,s)=>s.render("institution/register",{title:"Register institution",user:r.session.user}));
app.get("/institution/purchase",       auth("institution"),(r,s)=>s.render("institution/purchase",{title:"Purchase credits",user:r.session.user}));
app.get("/institution/user-management",auth("institution"),(r,s)=>s.render("institution/userManagement",{title:"User management",user:r.session.user}));

/* ---------- start server ---------- */
const PORT = process.env.PORT || 3001;
app.listen(PORT, () => console.log(`✔ http://localhost:${PORT}`));
