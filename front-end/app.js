// front-end/app.js
import express           from 'express';
import session           from 'express-session';
import path              from 'path';
import { fileURLToPath } from 'url';
import morgan            from 'morgan';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const app       = express();
const API_BASE = process.env.ORCHESTRATOR_URL || 'http://orchestrator:8080';

/*───────────────────────────
  1) 3rd-party middleware
───────────────────────────*/
app.use(morgan('dev'));

/*───────────────────────────
  2) Static assets
───────────────────────────*/
app.use(express.static(path.join(__dirname, 'public')));

/*───────────────────────────
  3) Body-parsers  (moved ↑)
     – form POSTs need req.body
───────────────────────────*/
app.use(express.urlencoded({ extended: false }));
app.use(express.json());

/*───────────────────────────
  4) Sessions  (+ locals helper)
───────────────────────────*/
app.use(session({
  secret           : 'change-this-secret',
  resave           : false,
  saveUninitialized: true,
}));
app.use((req, res, next) => {
  res.locals.user       = req.session.user || null;
  res.locals.currentUrl = req.originalUrl;
  next();
});

/*───────────────────────────
  5) EJS templating
───────────────────────────*/
app.set('view engine', 'ejs');
app.set('views', path.join(__dirname, 'views'));

/*───────────────────────────
  6) Auth helper
───────────────────────────*/
function auth(role) {
  return (req, res, next) => {
    if (!req.session.user) return res.redirect('/login');
    if (role) {
      // accept institution + representative aliases
      if (role === 'institution') {
        if (!['institution', 'representative', 'institution_representative']
              .includes(req.session.user.role)) {
          return res.redirect(`/${req.session.user.role}`);
        }
      } else if (req.session.user.role !== role) {
        return res.redirect(`/${req.session.user.role}`);
      }
    }
    next();
  };
}

/*───────────────────────────
  7) Dummy users (dev only)
───────────────────────────*/
const users = { alice: 'student', bob: 'instructor', iris: 'institution' };

/*───────────────────────────
  8) UI routes
───────────────────────────*/

// Home
app.get('/', (req, res) => {
  if (!req.session.user) return res.redirect('/login');
  res.redirect(`/${req.session.user.role}`);
});

// Signup / Login ────────────────────────────────────────────
app.get('/signup', (_, res) =>
  res.render('signup', { title: 'Sign Up', user: null })
);

app.get('/login', (_, res) =>
  res.render('login', { title: 'Log in', error: null, user: null })
);

// CLASSIC form POST – creates session
app.post('/login', async (req, res) => {
  const { username, password } = req.body;
  try {
    // Call orchestrator → user-management service
    const response = await fetch('${API_BASE}/user/login', {
      method : 'POST',
      headers: { 'Content-Type': 'application/json' },
      body   : JSON.stringify({ username, password })
    });
    const data = await response.json();

    if (!response.ok || !data.role) {
      return res.render('login', {
        title : 'Log in',
        error : data.message || 'Invalid credentials',
        user  : null
      });
    }

    // ⚠️  SAVE USER TO SESSION
    req.session.user = { username, role: data.role };

    // Redirect by role
    if (['institution_representative', 'representative']
        .includes(data.role))      return res.redirect('/institution');
    else if (data.role === 'instructor') return res.redirect('/instructor');
    else if (data.role === 'student')    return res.redirect('/student');
    else                                 return res.redirect('/');
  } catch (err) {
    return res.render('login', {
      title : 'Log in',
      error : 'Login failed',
      user  : null
    });
  }
});
app.post('/api/session', (req, res) => {
  const { username, role } = req.body;
  if (!username || !role) {
    return res.status(400).json({ error: 'username and role required' });
  }
  req.session.user = { username, role };
  res.sendStatus(200);
});
app.get('/logout', (req, res) =>
  req.session.destroy(() => res.redirect('/login'))
);

// Student UI ────────────────────────────────────────────────
app.get('/student',            auth('student'), (req,res)=>res.render('student/dashboard',    { user:req.session.user, title:'Dashboard' }));
app.get('/student/statistics', auth('student'), (req,res)=>res.render('student/statistics',   { user:req.session.user, title:'Statistics' }));
app.get('/student/my-courses', auth('student'), (req,res)=>res.render('student/myCourses',    { user:req.session.user, title:'My Courses' }));
app.get('/student/request',    auth('student'), (req,res)=>res.render('student/reviewRequest',{ user:req.session.user, title:'Review Request' }));
app.get('/student/status',     auth('student'), (req,res)=>res.render('student/reviewStatus', { user:req.session.user, title:'Review Status' }));
app.get('/student/personal',   auth('student'), (req,res)=>res.render('student/personal',     { user:req.session.user, title:'Personal Grades' }));

// Instructor UI ─────────────────────────────────────────────
app.get('/instructor',              auth('instructor'), (req,res)=>res.render('instructor/dashboard', { user:req.session.user, title:'Dashboard' }));
app.get('/instructor/post-initial', auth('instructor'), (req,res)=>res.render('instructor/postInitial',{ user:req.session.user, title:'Post Initial' }));
app.get('/instructor/post-final',   auth('instructor'), (req,res)=>res.render('instructor/postFinal',  { user:req.session.user, title:'Post Final' }));
app.get('/instructor/review-list',  auth('instructor'), (req,res)=>res.render('instructor/reviewList', { user:req.session.user, title:'Review Requests' }));
app.get('/instructor/reply',        auth('instructor'), (req,res)=>{
  const request_id = req.query.req || '';
  res.render('instructor/replyForm',{
    user        : req.session.user,
    title       : 'Reply to Review Request',
    request_id,
    course_name : 'software II',
    exam_period : 'spring 2025',
    student_name: 'john doe',
  });
});
app.get('/instructor/statistics',   auth('instructor'), (req,res)=>res.render('instructor/statistics', { user:req.session.user, title:'Statistics' }));

// Institution UI ────────────────────────────────────────────
app.get('/institution',                 auth('institution'), (req,res)=>res.render('institution/dashboard',      { user:req.session.user, title:'Dashboard' }));
app.get('/institution/register',        auth('institution'), (req,res)=>res.render('institution/register',       { user:req.session.user, title:'Register' }));
app.get('/institution/purchase',        auth('institution'), (req,res)=>res.render('institution/purchase',       { user:req.session.user, title:'Purchase' }));
app.get('/institution/user-management', auth('institution'), (req,res)=>res.render('institution/userManagement', { user:req.session.user, title:'Users' }));
app.get('/institution/statistics',      auth('institution'), (req,res)=>res.render('institution/statistics',     { user:req.session.user, title:'Statistics' }));

/*───────────────────────────
  9) Server start-up
───────────────────────────*/
const PORT = process.env.PORT || 3000;
app.listen(PORT, () =>
  console.log(`✔ Front-end listening at http://localhost:${PORT}`)
);
