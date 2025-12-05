"use client"

import React, { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import Image from 'next/image'
import { setCurrentUser } from '../utils/storage'
import './styles.css'

type User = { id: number; name: string; email: string; password: string }

export const nombres: string[] = [
  "Mateo", "Sofía", "Sebastián", "Valentina", "Diego", "Isabella", "Lucas", "Camila",
  "Benjamín", "Mariana", "Alejandro", "Gabriela", "Daniel", "Nicole", "Samuel", "Emma",
  "Adrián", "Renata", "Thiago", "Lucía", "Julián", "Antonella", "Andrés", "Sara",
  "Leonardo", "Martina", "Bruno", "Paula", "Gael", "Alexa", "Javier", "Mía",
  "Santiago", "Carla", "Emiliano", "Danna", "Iván", "Regina", "Fernando", "Josefina",
  "Hugo", "Ana", "Marco", "Clara", "Jorge", "Romina", "Pablo", "Victoria",
  "Raúl", "Elena", "Simón", "Laura", "Marcelo", "Julia", "Fabián", "Pilar",
  "Cristian", "Alicia", "Óscar", "Amelia", "Alexis", "Bianca", "Óliver", "Cecilia",
  "Rafael", "Daniela", "Esteban", "Carolina", "Max", "Florencia", "Gabriel", "Tania",
  "Lorenzo", "Noelia", "Erick", "Isidora", "David", "Ariana", "Mauricio", "Brenda",
  "Gonzalo", "Mónica", "Alan", "Elsa", "Liam", "Miranda", "Ethan", "Michelle",
  "Cristóbal", "Ivanna", "Pedro", "Melanie", "Héctor", "Bárbara", "Tomás", "Alejandra",
  "Ángel", "Natalia", "Ismael", "Lorena", "Rodolfo", "Tamara", "Kevin", "Zoe",
  "Mario", "Araceli", "Nelson", "Carmina", "Walter", "Rebeca", "Nicolás", "Patricia",
  "Facundo", "Lidia", "Alberto", "Mabel", "Rubén", "Estela", "Eduardo", "Claudia",
  "César", "Denisse", "Guillermo", "Ángela", "Félix", "Teresa", "Armando", "Silvia",
  "Ricardo", "Luna", "Saúl", "Nadia", "Elías", "Pamela", "Ómar", "Mara",
  "Tobías", "Adela", "Aarón", "Jimena", "Benicio", "Marisol", "Damián", "Lucero",
  "Emanuel", "Roxana", "Kevin", "Salma", "Gustavo", "Emilia", "Bastián", "Verónica",
  "Rodrigo", "Helena", "Jhonatan", "Camila", "Cristian", "Abigail", "Darío", "Carolina",
  "Jasson", "Scarlett", "Lisandro", "Kiara", "Renzo", "Aitana", "Elmer", "Ariadna",
  "Dante", "María José", "Iker", "Catalina", "Fabricio", "Jazmín", "Matías", "Nicole",
  "Franco", "Alina", "Ramiro", "Guadalupe", "Santino", "Mariana", "Agustín", "Lola",
  "Ignacio", "Celeste", "Eric", "Magdalena", "Abel", "Mirella", "Ezequiel", "Adriana",
  "Jonathan", "Atenea", "Aaron", "Clarisa", "Joshua", "Belén", "Anthony", "Elizabeth",
  "Jhon", "Mireya", "Michael", "Gloria", "Brian", "Rosalía", "Kevin", "Fátima"
];

function randomName(id: number) {

  const nombre = localStorage.getItem(`usuario${id}_data`);
  if (nombre) {
    const parsed = JSON.parse(nombre);
    return parsed.name;
  }

  const pos = Math.floor(Math.random() * nombres.length);
  return nombres[pos]
}

function randomEmail(id: number) {
  return `usuario${id}@example.com`
}

function randomPassword(id: number) {
  const usuario = localStorage.getItem(`usuario${id}_data`);
  if (usuario) {
    const parsed = JSON.parse(usuario);
    return parsed.password;
  }

  return Math.random().toString(36).slice(2, 10)
}

function generarColorRGB() {
  // Generar un número entre 0 y 255 para cada canal
  const r = Math.floor(Math.random() * 256);
  const g = Math.floor(Math.random() * 256);
  const b = Math.floor(Math.random() * 256);

  // Retornar la cadena con formato CSS
  return `rgb(${r}, ${g}, ${b})`;
}

export default function UserSelection() {
  const [ids, setIds] = useState<number[]>([])
  const [loading, setLoading] = useState(true)
  const router = useRouter()

  // Estados nuevos para el modal de login
  const [selectedUserId, setSelectedUserId] = useState<number | null>(null)
  const [inputEmail, setInputEmail] = useState<string>("")
  const [inputPassword, setInputPassword] = useState<string>("")
  const [loginError, setLoginError] = useState<string>("")

  useEffect(() => {
    fetch('/api/users')
      .then((r) => r.json())
      .then((data) => {
        // API returns array of ids (strings or numbers)
        const all = (data || []).map((x: string | number) => Number(x))
        // shuffle and pick up to 20 random users to avoid overloading the UI
        const shuffled = [...all].sort(() => 0.5 - Math.random())
        const list = shuffled.slice(0, 20)
        setIds(list)
        // Ensure stable user records in localStorage so displayed name
        // matches the one stored when opening the modal.
        try {
          list.forEach((id) => {
            const key = `usuario${id}_data`
            if (!localStorage.getItem(key)) {
              const userToStore = {
                id,
                name: nombres[Math.floor(Math.random() * nombres.length)],
                email: randomEmail(id),
                password: Math.random().toString(36).slice(2, 10)
              }
              try {
                localStorage.setItem(key, JSON.stringify(userToStore))
              } catch (err) {
                // ignore quota errors, fallback will still work
                console.error('No se pudo crear usuario en localStorage', err)
              }
            }
          })
        } catch (err) {
          console.error('Error al inicializar usuarios en localStorage', err)
        }
      })
      .catch(() => setIds([]))
      .finally(() => setLoading(false))
  }, [])

  // Al elegir usuario, abrimos el modal en lugar de navegar directamente
  function chooseUser(id: number) {
    // Si no existe, crear registro por defecto en localStorage antes de abrir el modal
    const key = `usuario${id}_data`
    const stored = localStorage.getItem(key)
    if (!stored) {
      const userToStore = {
        id,
        name: randomName(id),
        email: randomEmail(id),
        password: randomPassword(id)
      }
      try {
        localStorage.setItem(key, JSON.stringify(userToStore))
      } catch (err) {
        console.error('No se pudo crear usuario en localStorage', err)
      }
    }

    // No rellenar el correo automáticamente: dejar vacío para que el usuario lo escriba
    setInputEmail('')
    setInputPassword('')
    setLoginError('')
    setSelectedUserId(id)
  }

  // Regenerar usuarios: limpiar y recrear registros en localStorage
  async function regenerateUsers() {
    setLoading(true)
    try {
      const r = await fetch('/api/users')
      const data = await r.json()
      const all = (data || []).map((x: string | number) => Number(x))
      const shuffled = [...all].sort(() => 0.5 - Math.random())
      const list = shuffled.slice(0, 20)
      // Recreate records for the selected ids
      try {
        list.forEach((id) => {
          const key = `usuario${id}_data`
          const userToStore = {
            id,
            name: nombres[Math.floor(Math.random() * nombres.length)],
            email: randomEmail(id),
            password: Math.random().toString(36).slice(2, 10)
          }
          try {
            localStorage.setItem(key, JSON.stringify(userToStore))
          } catch (err) {
            console.error('No se pudo crear usuario en localStorage', err)
          }
        })
      } catch (err) {
        console.error('Error al regenerar usuarios en localStorage', err)
      }
      setIds(list)
    } catch (err) {
      console.error('Error al regenerar usuarios', err)
    } finally {
      setLoading(false)
    }
  }

  // Verificar credenciales cuando el usuario envía el formulario
  function handleLogin() {
    if (selectedUserId === null) return
    const key = `usuario${selectedUserId}_data`
    const stored = localStorage.getItem(key)
    if (!stored) {
      // Fallback: si por alguna razón no existe, comprobar con randomPassword
      const expectedPassword = randomPassword(selectedUserId)
      if (inputPassword === expectedPassword) {
        const user: User = { id: selectedUserId, name: randomName(selectedUserId), email: inputEmail, password: inputPassword }
        setCurrentUser(user)
        router.push('/menu')
        return
      }
      setLoginError('Error al iniciar sesión: credenciales incorrectas')
      return
    }

    try {
      const parsed = JSON.parse(stored)
      const storedEmail = parsed.email
      const storedPassword = parsed.password

      // Si no hay email almacenado, aceptar si la contraseña coincide y guardar el email introducido
      if (!storedEmail) {
        if (inputPassword === storedPassword) {
          // actualizar email en localStorage
          parsed.email = inputEmail
          try {
            localStorage.setItem(key, JSON.stringify(parsed))
          } catch (err) {
            console.error('No se pudo actualizar email en localStorage', err)
          }
          const user: User = { id: selectedUserId, name: parsed.name || randomName(selectedUserId), email: inputEmail, password: storedPassword }
          setCurrentUser(user)
          router.push('/menu')
          return
        }
        setLoginError('Error al iniciar sesión: credenciales incorrectas')
        return
      }

      if (inputEmail === storedEmail && inputPassword === storedPassword) {
        const user: User = { id: selectedUserId, name: parsed.name || randomName(selectedUserId), email: storedEmail, password: storedPassword }
        setCurrentUser(user)
        router.push('/menu')
      } else {
        setLoginError('Error al iniciar sesión: credenciales incorrectas')
      }
    } catch (_e) {
      console.error('parse error usuario data', _e)
      setLoginError('Error al iniciar sesión: datos corruptos')
    }
  }

  function cancelLogin() {
    setSelectedUserId(null)
    setLoginError('')
  }

  return (
    <main>
      <div className="container-fluid m-0 contenedor-pagina-inicio">
        <div className="row justify-content-center align-items-center">

          <div className="col-12">
            <div className="container-fluid">
              <div className="row justify-content-start align-items-center">
                <div className="col-12 text-center">
                  <h1 className="titulo-bienvenida">Bienvenido a NextWatch</h1>
                </div>
                <div className="col-12 text-center">
                  <h1 className='titulo-elegir'>Elige un usuario</h1>
                  <div style={{ marginTop: 8 }}>
                    <button onClick={() => regenerateUsers()} style={{ padding: '6px 10px', fontSize: 14, color: 'white', backgroundColor: '#007bff', border: 'none', borderRadius: 4 }}>Regenerar usuarios</button>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div className="col-12">
            <div className="container-fluid">
              <div className="row justify-content-center align-items-center">

                <div className="col-12 text-center">
                  {loading && <div>Cargando usuarios...</div>}
                  {!loading && ids.length === 0 && <div>No hay usuarios (comprueba /api/users)</div>}
                </div>

                <div className="col-12 text-center">

                  <div className="container-fluid p-5">
                    <div className="row justify-content-center align-items-center">
                        {ids.map((id) => (
                          <div className='col-4' key={id} style={{ margin: 8 }}>
                            <button onClick={() => chooseUser(id)}>

                              <div className="container-fluid contenedor-usuario p-3" style={{ backgroundColor: generarColorRGB() }}>
                                <div className="row justify-content-center align-items-center">
                                  <div className="col-12 contenedor-icono-user text-center">
                                    <Image src="/user.png" alt="User Icon" className="user-icon img-fluid" width={96} height={96} />
                                  </div>
                                  <div className="col-12">
                                    <h3 className='texto-nombre'>Nombre: {randomName(id)}</h3>
                                  </div>
                                </div>
                              </div>
                            </button>
                          </div>
                        ))}
                    </div>
                  </div>
                </div>

              </div>
            </div>
          </div>

        </div>
      </div>

      {/* Modal sencillo para pedir email y contraseña */}
      {selectedUserId !== null && (
        <div className="login-modal-overlay" style={{ position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, background: 'rgba(0,0,0,0.5)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 9999 }}>
          <div className="login-modal" style={{ background: '#fff', padding: 24, borderRadius: 8, width: 360, maxWidth: '90%' }}>
            <h3>Iniciar sesión - Usuario {selectedUserId}</h3>
            <div style={{ marginTop: 8 }}>
              <label style={{ display: 'block', fontSize: 14 }}>Correo</label>
              <input value={inputEmail} onChange={(e) => setInputEmail(e.target.value)} style={{ width: '100%', padding: 8, marginTop: 4 }} />
            </div>
            <div style={{ marginTop: 8 }}>
              <label style={{ display: 'block', fontSize: 14 }}>Contraseña</label>
              <input type="password" value={inputPassword} onChange={(e) => setInputPassword(e.target.value)} style={{ width: '100%', padding: 8, marginTop: 4 }} />
            </div>
            {loginError && <div style={{ color: 'red', marginTop: 8 }}>{loginError}</div>}
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 8, marginTop: 12 }}>
              <button onClick={cancelLogin} style={{ padding: '8px 12px' }}>Cancelar</button>
              <button onClick={handleLogin} style={{ padding: '8px 12px' }}>Iniciar sesión</button>
            </div>
          </div>
        </div>
      )}
    </main>
  )
}
